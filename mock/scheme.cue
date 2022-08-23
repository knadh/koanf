{
	#labels: [string]:	string
	#labels: app:		string

	#metadata: {
		name:		string
		namespace?:	string
		labels:		#labels
		annotations?:	[string]: string
	}
		
	{
		apiVersion: "apps/v1"
		kind: "Deployment"
		metadata:	#metadata
		spec: {
			selector: {
				matchLabels: metadata.labels
			}
			strategy: {...}
			minReadySeconds: uint
			template: {
				metadata: {...}
				spec: {
					containers: [...{
						...
						ports: [...{
							containerPort: 80
							protocol: "TCP"
						}]
					}]
				}
			}
		}
		...
	}
}
