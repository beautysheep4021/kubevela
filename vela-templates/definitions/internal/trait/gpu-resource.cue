"gpu-resource": {
	type: "trait"
	annotations: {}
	description: "Adds GPU requests and limits to AI workloads without introducing AI-specific top-level resources."
	attributes: {
		podDisruptive:      true
		appliesToWorkloads: ["deployments.apps", "jobs.batch"]
	}
}
template: {
	patch: {
		let gpuContent = {
			resources: {
				requests: {
					"nvidia.com/gpu": parameter.count
				}
				limits: {
					"nvidia.com/gpu": parameter.count
				}
			}
		}

		if context.output.spec != _|_ if context.output.spec.template != _|_ {
			spec: template: spec: {
				// +patchKey=name
				containers: [gpuContent]
				if parameter.nodeSelector != _|_ {
					nodeSelector: parameter.nodeSelector
				}
				if parameter.tolerations != _|_ {
					tolerations: parameter.tolerations
				}
			}
		}
	}

	parameter: {
		// +usage=Number of GPUs requested by the workload
		count: *1 | int & >=1
		// +usage=Node selector for GPU-capable nodes
		nodeSelector?: [string]: string
		// +usage=Tolerations for GPU node taints
		tolerations?: [...{
			key?:               string
			operator?:          *"Equal" | "Equal" | "Exists"
			value?:             string
			effect?:            string
			tolerationSeconds?: int
		}]
	}
}
