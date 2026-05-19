"ai-runtime": {
	type: "trait"
	annotations: {}
	description: "Adds AI runtime metadata to workload pod templates for observability, audit, and later policy integration."
	attributes: {
		appliesToWorkloads: ["deployments.apps", "jobs.batch"]
	}
}
template: {
	patch: {
		let runtimeLabels = {
			"ai.oam.dev/runtime": parameter.runtime
			if parameter.framework != _|_ {
				"ai.oam.dev/framework": parameter.framework
			}
			if parameter.tenant != _|_ {
				"ai.oam.dev/tenant": parameter.tenant
			}
			if parameter.project != _|_ {
				"ai.oam.dev/project": parameter.project
			}
			if parameter.environment != _|_ {
				"ai.oam.dev/environment": parameter.environment
			}
		}
		let runtimeAnnotations = {
			if parameter.modelURI != _|_ {
				"ai.oam.dev/model-uri": parameter.modelURI
			}
			if parameter.datasetURI != _|_ {
				"ai.oam.dev/dataset-uri": parameter.datasetURI
			}
			if parameter.owner != _|_ {
				"ai.oam.dev/owner": parameter.owner
			}
		}

		if context.output.spec != _|_ if context.output.spec.template != _|_ {
			spec: template: metadata: {
				labels: runtimeLabels
				annotations: runtimeAnnotations
			}
		}
	}

	parameter: {
		// +usage=AI runtime name, such as triton, vllm, pytorch, tensorflow, or custom
		runtime: string
		// +usage=AI framework name
		framework?: string
		// +usage=Tenant that owns the AI workload
		tenant?: string
		// +usage=Project that owns the AI workload
		project?: string
		// +usage=Environment where the AI workload runs
		environment?: string
		// +usage=Model artifact URI for audit and observability
		modelURI?: string
		// +usage=Dataset URI for audit and observability
		datasetURI?: string
		// +usage=Owning team or project
		owner?: string
	}
}
