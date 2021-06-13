muchk8s.wow

Requirements: The following need to be installed
- kubectl
- helm
- kind installed


To start the env:
`make create`

This will create the deployment of 3 namespaces/
Alice - protected-b
Bob - unclassidied
carla - no label of classificaiton

To delete the env:
`make delete`


Issues:
If the deployments don't start because of insufficient CPUs.
`kubectl edit deployment gatekeeper-controller-manager -n gatekeeper-system`
Go to the line about resources, delete the lines and have `resources: {}`

Need to apply the `uploadPolicy.yaml` to alice
