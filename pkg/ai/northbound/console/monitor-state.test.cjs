const { test } = require("node:test");
const assert = require("node:assert/strict");
require("./state.js");
require("./monitor-state.js");
const S = globalThis.MonitorState;
test("informational phase and component messages retain info severity", () => {
  const evidence = S.diagnostics({
    phase: "running",
    message: "Reconciled",
    components: [{ name: "worker", message: "Ready" }],
  });
  assert.equal(evidence.length, 2);
  assert.ok(evidence.every((d) => d.severity === "info"));
  assert.equal(
    S.diagnostics({ phase: "failed", message: "exit 1" })[0].severity,
    "error",
  );
});
test("failed detail returns to original list with namespace and status filters", () => {
  const detail = S.route(
    "#/jobs/a/missing?from=jobs&ns=a&status=failed&q=train",
  );
  assert.equal(
    S.hash({ ...detail, detail: false, section: detail.from }),
    "#/jobs?q=train&ns=a&status=failed",
  );
});
test("diagnostic summary prioritizes strongest evidence over repeated phase messages", () => {
  const item = {
    phase: "pending",
    message: "FailedScheduling: no CPU",
    diagnostics: [
      { severity: "warning", reason: "FailedScheduling", message: "no CPU" },
    ],
  };
  assert.equal(S.diagnosticSummary(item).reason, "FailedScheduling");
  assert.equal(S.diagnosticSummary({ healthy: false }), null);
});
test("request display converts exact CPU and memory units without inventing missing values", () => {
  assert.deepEqual(S.requestParts({ cpuMilli: 1000, memoryMi: 1024, gpu: 0 }), [
    "1 CPU",
    "1 GiB",
    "0 GPU",
  ]);
  assert.deepEqual(S.requestParts({ cpuMilli: 500, memoryMi: 512 }), [
    "0.5 CPU",
    "512 MiB",
    "0 GPU",
  ]);
  assert.equal(S.requestParts(null), null);
});
test("namespaced routes and filters round trip without conflating target and scope", () => {
  const r = S.route(
    "#/jobs/team-a/same?tab=logs&ns=team-b&q=a%26b&tenant=t&status=pending",
  );
  assert.equal(r.namespace, "team-a");
  assert.equal(r.ns, "team-b");
  assert.equal(r.q, "a&b");
  assert.deepEqual(S.route(S.hash(r)), r);
  assert.equal(S.route("#/applications/ns/name").detail, true);
  assert.equal(S.route("#/jobs/only-ns").invalid, true);
  assert.equal(S.route("#/jobs/%ZZ/name").invalid, true);
});
test("detail and return links retain list filters and audit origin", () => {
  const origin = S.route("#/audits?actor=admin&name=job&outcome=failure&ns=a");
  const target = S.route(S.detailHash({ namespace: "a", name: "job" }, origin));
  assert.equal(target.actor, "admin");
  assert.equal(target.objectName, "job");
  assert.equal(target.from, "audits");
  assert.equal(
    S.hash({ ...target, section: target.from, detail: false }),
    "#/audits?ns=a&actor=admin&outcome=failure&name=job",
  );
});
test("filters combine namespace tenant name and shared status", () => {
  const rows = [
    {
      name: "same",
      namespace: "a",
      kind: "AIJob",
      phase: "pending",
      aiMetadata: { "ai.oam.dev/tenant": "one" },
    },
    { name: "same", namespace: "b", kind: "AIJob", phase: "failed" },
  ];
  assert.equal(
    S.filter(rows, { ns: "a", tenant: "one", q: "SAM", status: "pending" })
      .length,
    1,
  );
  assert.equal(S.filter(rows, { ns: "b", tenant: "one" }).length, 0);
});
test("audit filters use explicit success and preserve unknown type for resolution", () => {
  const row = {
    actor: "alice",
    action: "rerun",
    success: false,
    name: "job",
    namespace: "a",
  };
  assert.equal(
    S.filterAudits([row], {
      actor: "ali",
      action: "rerun",
      outcome: "failure",
      name: "jo",
      ns: "a",
    }).length,
    1,
  );
  assert.equal(S.filterAudits([row], { outcome: "success" }).length, 0);
  assert.equal(S.detailHash(row), "#/applications/a/job");
});
test("resource grouping separates tenants namespaces and completed declarations", () => {
  const row = {
    kind: "AIJob",
    name: "j",
    namespace: "a",
    phase: "succeeded",
    resourceSummary: { cpuMilli: 500, memoryMi: 512, gpu: 1 },
    aiMetadata: { "ai.oam.dev/tenant": "t" },
  };
  const groups = S.groupResources([
    row,
    { ...row, phase: "pending" },
    { ...row, namespace: "b" },
  ]);
  assert.equal(groups.length, 2);
  assert.equal(groups[0].completed.count, 1);
  assert.equal(groups[0].other.cpuMilli, 500);
  assert.equal(groups[0].completed.gpu, 1);
  assert.equal(
    S.groupResources([{ ...row, resourceSummary: undefined }])[0].completed
      .missing,
    1,
  );
});
test("quota rows retain separate quotas and missing values", () => {
  const rows = S.quotaRows([
    {
      namespace: "a",
      quotas: [
        { name: "one", hard: { cpu: "2" }, used: { cpu: "0" } },
        { name: "two", hard: { cpu: "4" }, used: {} },
      ],
    },
  ]);
  assert.equal(rows.length, 2);
  assert.equal(rows[0].used, "0");
  assert.equal(rows[1].used, null);
});
test("quota desired caps do not masquerade as enforced or consumed quota", () => {
  const rows = S.quotaRows([
    {
      namespace: "a",
      quotas: [
        { name: "pending", desiredHard: { cpu: "4" }, hard: {}, used: {} },
      ],
    },
  ]);
  assert.equal(rows.length, 1);
  assert.equal(rows[0].desiredHard, "4");
  assert.equal(rows[0].hard, null);
  assert.equal(rows[0].used, null);
});
test("abnormal evidence excludes healthy false and completed informational evidence", () => {
  assert.deepEqual(S.diagnostics({ healthy: false, phase: "pending" }), []);
  assert.deepEqual(
    S.diagnostics({
      phase: "succeeded",
      diagnostics: [{ severity: "info", reason: "Completed" }],
    }),
    [],
  );
  assert.equal(
    S.diagnostics({ phase: "failed", message: "exit 1" })[0].message,
    "exit 1",
  );
  const item = {
    components: [
      {
        name: "c",
        pods: [
          {
            name: "p",
            events: [
              {
                type: "Warning",
                reason: "FailedScheduling",
                message: "No GPU",
              },
            ],
            containers: [
              { name: "main", reason: "CrashLoopBackOff", message: "exit 2" },
            ],
          },
        ],
      },
    ],
  };
  assert.equal(S.diagnostics(item).length, 2);
  assert.ok(S.diagnostics(item).every((d) => d.resource.includes("p")));
});
