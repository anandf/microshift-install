# microshift-install

# Build steps
- Clone this repository
- Build the `microshift-install` binary for both `arm64` and `amd64` architectures using the following command.
```
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o microshift-install-linux-amd64 .
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o microshift-install-linux-arm64 .
```

# Installation steps
- Transfer the binary to the microshift instance
```
scp -i ~/.ssh/id_rsa microshift-install-linux-<arch> <user>@<host>:/home/<user>
```
eg:
```
scp -i ~/.ssh/id_rsa microshift-install-linux-amd64 ec2-user@18.237.171.200:/home/ec2-user
```
- Give execute permission for the binary and see if you are able to execute it.
```
chmod +x /home/<user>/microshift-install
/home/<user>/microshift-install --help
```
eg:
```
chmod +x /home/ec2-user/microshift-install
/home/ec2-user/microshift-install --help
```
- Run the binary to install the kustomize manifests in `/etc/microshift/manifests.d` directory
```
sudo /home/ec2-user/microshift-install install <image>
```
```
sudo /home/ec2-user/microshift-install install quay.io/anjoseph/openshift-gitops-1-gitops-microshift-bundle:v99.9.0-8
```
Note: The above command should be run with super user previlieges (sudo). The command would extract the `kustomize` manifests present in the OCI container image to `/etc/microshift/manifests.d/` directory and also restart the microshift systemd service

# Steps for initializing microshift instance created via cluster-bot workflow
- Set the KUBECONFIG environment variable
```
export KUBECONFIG=/path/to/downloaded/kubeconfig/from/clusterbot
```
- Get the SSH public key contents
eg:
```
cat ~/.ssh/id_rsa.pub
```
- Get the IP of the microshift instance from the KUBECONFIG file
```
MICROSHIFT_IP=$(cat $KUBECONFIG | grep server | sed -e 's/server: https:\/\///' | cut -f1 -d:)

- Get the node name by running the following command. There would be only a single node for the microshift instance
```
oc get nodes
```
```
- Login to the node by creating a node debugging session using the following command
```
oc debug node/<node-name>
```
eg:
```
oc debug node/i-02e1b93c92d3baffe.us-west-2.compute.internal
```
- Paste the public key of your workstation (from previous step) as authorized keys of the microshift instance and exit from the node debugging session.
```
echo <pub_key> >> /home/ec2-user/.ssh/authorized_keys
```
- SSH into the machine using the IP address calculated in the previous step `MICROSHIFT_IP` with the below command
```
ssh -i <path/to/private_key> <user>@<host_ip>
```
eg:
```
ssh -i ~/.ssh/id_rsa ec2-user@18.237.171.200
```
- Test if you are able to scp a file to the remote microshift instance
```
scp -i <path/to/private_key> <src_file> <user>@<host_ip>:<destination_dir>
```
eg:
```
scp -i ~/.ssh/id_rsa microshift-install-linux-amd64 ec2-user@18.237.171.200:/home/ec2-user/microshift-install
```
# Steps for initializing microshift instance created via OpenShift Local (CRC)

- Download the latest version of OpenShift Local (CRC) binary from the below link
```
https://developers.redhat.com/products/openshift-local/overview
```

- Setup the CRC cluster by running the following command
```
crc setup
```
- Set the preset to `microshift` using the following command
```
crc config set preset microshift
```
- Start the microshift instance
```
crc start
```
- Check if the microshift instance is running
```
crc status
```
- Get the IP address of the CRC based microshift instance
```
crc ip
```
- SSH into the microshift instance using the following command
```
ssh -i ~/.crc/machines/crc/id_ecdsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p <port> core@<crc_ip>
```
eg:
```
ssh -i ~/.crc/machines/crc/id_ecdsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p 2222 core@127.0.0.1
```

- Test if you are able to scp a file to the remote microshift instance
```
scp -i -i ~/.crc/machines/crc/id_ecdsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null <src_file> -P <port> core@<crc_ip>:<destination_dir>
```
eg:
```
scp -i ~/.crc/machines/crc/id_ecdsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null microshift-install-linux-amd64 -P 2222 crc@127.0.0.1:/home/ec2-user/microshift-install
```

# Running argocd in core mode

- Set the namespace in the context to `openshift-gitops`
```
kubectl config set-context --current --namespace openshift-gitops
argocd --core app list
```

# Creating an Argo Application using argocd CLI
```
argocd app create guestbook --core --repo https://github.com/anandf/argocd-example-apps.git --path guestbook/dev --dest-namespace guestbook --dest-server https://kubernetes.default.svc --directory-recurse --sync-policy automated --self-heal --sync-option CreateNamespace=true
```
```
argocd app create kserve --core --repo https://github.com/anandf/microshift-install.git --path kserve/base --dest-namespace kserve --dest-server https://kubernetes.default.svc --directory-recurse --sync-policy automated --self-heal --sync-option CreateNamespace=true
```