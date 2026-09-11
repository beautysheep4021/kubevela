(function (root) {
  "use strict";
  const S = root.ConsoleState;
  const filterKeys = [
    "q",
    "ns",
    "tenant",
    "status",
    "actor",
    "action",
    "outcome",
    "name",
  ];
  function route(hash) {
    try {
      const u = new URL(
        (hash || "#/overview").replace(/^#/, ""),
        "http://monitor",
      );
      const p = u.pathname.split("/").filter(Boolean).map(decodeURIComponent);
      const section = p[0] || "overview";
      const detail =
        ["jobs", "services", "applications"].includes(section) &&
        p.length === 3;
      const invalid =
        ![
          "overview",
          "jobs",
          "services",
          "tenants",
          "audits",
          "applications",
        ].includes(section) ||
        (p.length > 1 && !detail) ||
        (section === "applications" && !detail);
      const r = {
        section,
        detail,
        invalid,
        namespace: detail ? p[1] : "",
        name: detail ? p[2] : "",
        tab: ["overview", "logs", "result", "config"].includes(
          u.searchParams.get("tab"),
        )
          ? u.searchParams.get("tab")
          : "overview",
      };
      filterKeys.forEach((k) => {
        if (k !== "name") r[k] = u.searchParams.get(k) || "";
      });
      r.objectName = u.searchParams.get("name") || "";
      r.from = ["overview", "jobs", "services", "tenants", "audits"].includes(
        u.searchParams.get("from"),
      )
        ? u.searchParams.get("from")
        : "";
      return r;
    } catch (_) {
      return { section: "overview", invalid: true };
    }
  }
  function hash(r) {
    const params = new URLSearchParams();
    filterKeys.forEach((k) => {
      const v = k === "name" ? r.objectName : r[k];
      if (v) params.set(k, v);
    });
    if (r.detail && r.tab !== "overview") params.set("tab", r.tab);
    if (r.detail && r.from) params.set("from", r.from);
    return (
      "#/" +
      r.section +
      (r.detail
        ? "/" +
          encodeURIComponent(r.namespace) +
          "/" +
          encodeURIComponent(r.name)
        : "") +
      (params.size ? "?" + params : "")
    );
  }
  const tenant = (i) =>
    i.tenant || (i.aiMetadata || {})["ai.oam.dev/tenant"] || "";
  const includes = (value, query) =>
    String(value || "")
      .toLowerCase()
      .includes(String(query || "").toLowerCase());
  function filter(items, r) {
    return items.filter(
      (i) =>
        includes(i.name, r.q) &&
        (!r.ns || i.namespace === r.ns) &&
        (!r.tenant || tenant(i) === r.tenant) &&
        (!r.status || S.status(i, S.kind(i)).key === r.status),
    );
  }
  function filterAudits(items, r) {
    return items.filter(
      (i) =>
        includes(i.name, r.objectName || r.name || r.q) &&
        includes(i.actor, r.actor) &&
        (!r.ns || i.namespace === r.ns) &&
        (!r.action || i.action === r.action) &&
        (!r.outcome ||
          (r.outcome === "success" ? i.success === true : i.success === false)),
    );
  }
  function detailHash(i, origin) {
    const kind = S.kind(i);
    const section =
      kind === "AIJob"
        ? "jobs"
        : kind === "AIService"
          ? "services"
          : "applications";
    return hash({
      ...origin,
      section,
      detail: true,
      namespace: i.namespace,
      name: i.name,
      tab: "overview",
      from: origin?.section || "",
    });
  }
  function groupResources(items) {
    const groups = new Map();
    const totals = () => ({
      count: 0,
      cpuMilli: 0,
      memoryMi: 0,
      gpu: 0,
      missing: 0,
    });
    items
      .filter((i) => S.kind(i))
      .forEach((i) => {
        const t = tenant(i),
          key = JSON.stringify([t, i.namespace]);
        if (!groups.has(key))
          groups.set(key, {
            tenant: t,
            namespace: i.namespace,
            completed: totals(),
            other: totals(),
          });
        const total =
          groups.get(key)[
            S.status(i, S.kind(i)).key === "succeeded" ? "completed" : "other"
          ];
        total.count++;
        if (!i.resourceSummary) {
          total.missing++;
          return;
        }
        ["cpuMilli", "memoryMi", "gpu"].forEach((k) => {
          const n = Number(i.resourceSummary[k] || 0);
          if (Number.isFinite(n) && n >= 0) total[k] += n;
        });
      });
    return [...groups.values()].sort((a, b) =>
      (a.tenant + "\0" + a.namespace).localeCompare(
        b.tenant + "\0" + b.namespace,
      ),
    );
  }
  function quotaRows(items) {
    return items.flatMap((i) =>
      (i.quotas || []).flatMap((q) => {
        const resources = [
          ...new Set([
            ...Object.keys(q.desiredHard || {}),
            ...Object.keys(q.hard || {}),
            ...Object.keys(q.used || {}),
          ]),
        ].sort();
        return resources.map((resource) => ({
          namespace: i.namespace,
          name: q.name,
          resource,
          desiredHard: q.desiredHard?.[resource] ?? null,
          hard: q.hard?.[resource] ?? null,
          used: q.used?.[resource] ?? null,
        }));
      }),
    );
  }
  function diagnostics(item) {
    const rows = [];
    const add = (d) => {
      if (d.reason !== "Completed" && (d.message || d.reason || d.evidence))
        rows.push(d);
    };
    if (/^(failed|error|errorappconfig)$/i.test(item.phase || ""))
      add({
        severity: "error",
        reason: item.phase,
        message: item.message || "",
        resource: item.name,
      });
    else if (
      item.message &&
      !/^(succeeded|completed|complete)$/i.test(item.phase || "")
    )
      add({
        severity: "info",
        reason: item.phase || "Message",
        message: item.message,
        resource: item.name,
      });
    (item.diagnostics || []).filter((d) => d.severity !== "info").forEach(add);
    (item.warnings || []).forEach((d) => add({ ...d, severity: "warning" }));
    (item.components || []).forEach((c) => {
      if (c.message)
        add({ severity: "info", resource: c.name, message: c.message });
      (c.pods || []).forEach((p) => {
        (p.events || [])
          .filter((e) => e.type === "Warning")
          .forEach((e) =>
            add({ ...e, severity: "warning", resource: c.name + "/" + p.name }),
          );
        (p.containers || [])
          .filter((k) => k.reason && k.reason !== "Completed")
          .forEach((k) =>
            add({
              severity: "warning",
              reason: k.reason,
              message: k.message,
              resource: c.name + "/" + p.name + "/" + k.name,
            }),
          );
      });
    });
    return rows.filter(
      (d, index) =>
        rows.findIndex(
          (x) =>
            x.resource === d.resource &&
            x.reason === d.reason &&
            x.message === d.message,
        ) === index,
    );
  }
  function diagnosticSummary(item) {
    const rank = { error: 3, warning: 2, info: 1 };
    return (
      diagnostics(item).sort(
        (a, b) => (rank[b.severity] || 0) - (rank[a.severity] || 0),
      )[0] || null
    );
  }
  function requestParts(value) {
    if (!value) return null;
    const memory = value.memoryMi || 0;
    return [
      `${(value.cpuMilli || 0) / 1000} CPU`,
      memory && memory % 1024 === 0 ? `${memory / 1024} GiB` : `${memory} MiB`,
      `${value.gpu || 0} GPU`,
    ];
  }
  root.MonitorState = {
    route,
    hash,
    filter,
    filterAudits,
    detailHash,
    tenant,
    groupResources,
    quotaRows,
    diagnostics,
    diagnosticSummary,
    requestParts,
  };
})(globalThis);
