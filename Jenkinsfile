pipeline {
  agent none
  stages {
    stage("Pre-build") {
      agent any
      steps {
        slackSend color: "#439FE0", message: "Build Started: ${env.JOB_NAME} ${env.BUILD_NUMBER} (${env.BUILD_URL})"
        // necessarry to grant jenkins the mounted docker socket
        // in order to spin up more docker containers below
        sh "sudo change-socket.docker"
      }
    }
    stage("Build") {
      agent {
        docker {
          image 'golang:1.13'
          args '-u root -v go-cache:/go'
          reuseNode true
        }
      }
      steps {
        sh "go build -o main && rm main"
      }
    }
    stage("Lint") {
      agent {
        docker {
          image 'golang:1.13'
          args '-u root -v go-cache:/go'
          reuseNode true
        }
      }
      steps {
        sh "go get golang.org/x/lint/golint"
        sh "$GOPATH/bin/golint ./..."
        slackSend color: "good", message: "Build Succeeded: ${env.JOB_NAME} ${env.BUILD_NUMBER} (${env.BUILD_URL})"
      }
    }
    stage("Deploy") {
      agent any
      when {
        branch "master"
      }
      steps {
        slackSend color: "#439FE0", message: "Deploy Started: ${env.JOB_NAME} ${env.BUILD_NUMBER} (${env.BUILD_URL})"
        sh 'sudo change-socket.docker'
        sh 'sudo rsync -r . /opt/slack-backer-upper --delete'
        sh 'cd /opt/telemetry-server && docker-compose -f docker-compose.yml build --pull'
        sh 'cd /opt/telemetry-server && docker-compose -f docker-compose.yml up -d --force-recreate slack-backer-upper'
        slackSend color: "good", message: "Deploy Succeeded: ${env.JOB_NAME} ${env.BUILD_NUMBER} (${env.BUILD_URL})"
      }
    }
  }
  post {
    // triggered when red sign
    failure {
      slackSend color: "danger", message: "Job Failed: ${env.JOB_NAME} ${env.BUILD_NUMBER} (${env.BUILD_URL})"
    }
  }
}
