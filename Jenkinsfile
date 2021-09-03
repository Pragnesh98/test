 def label = "worker-${UUID.randomUUID().toString()}"

 podTemplate(label: label, containers: [
   containerTemplate(name: 'docker', image: 'docker', command: 'cat', ttyEnabled: true),
   containerTemplate(name: 'kubectl', image: 'lachlanevenson/k8s-kubectl:v1.17.2', command: 'cat', ttyEnabled: true),
 ],
 volumes: [
   hostPathVolume(mountPath: '/var/run/docker.sock', hostPath: '/var/run/docker.sock')
 ]) {
   node(label) {
     def myRepo = checkout scm
     def gitCommit = myRepo.GIT_COMMIT
     def gitBranch = myRepo.GIT_BRANCH
     def shortGitCommit = "${gitCommit[0..10]}"
     def previousGitCommit = sh(script: "git rev-parse ${gitCommit}~", returnStdout: true)

     stage('Create Docker images') {
       container('docker') {
         withCredentials([[$class: 'UsernamePasswordMultiBinding',
           credentialsId: 'azurecr',
           usernameVariable: 'DOCKER_HUB_USER',
           passwordVariable: 'DOCKER_HUB_PASSWORD']]) {
           sh """
             docker login yellowmessenger.azurecr.io -u ${DOCKER_HUB_USER} -p ${DOCKER_HUB_PASSWORD}
             docker build -t yellowmessenger.azurecr.io/asterisk-ari:v${BUILD_NUMBER} .
             docker push yellowmessenger.azurecr.io/asterisk-ari:v${BUILD_NUMBER}
             """
         }
       }
     }
   }
 }
