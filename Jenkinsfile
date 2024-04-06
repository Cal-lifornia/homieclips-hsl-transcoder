pipeline {
    // install golang 1.14 on Jenkins node
    agent any
    tools {
        go 'go1.21.8'
    }
    environment {
        CGO_ENABLED = 0
        GOPATH = "${JENKINS_HOME}/jobs/${JOB_NAME}/builds/${BUILD_ID}"
    }
    stages {
        stage("unit-test") {
            steps {
                echo 'UNIT TEST EXECUTION STARTED'
                sh 'go test ./...'
            }
        }
        stage("build") {
            agent{
                docker {
                    image 'docker:24.0.5'
                    reuseNode true
                }
            }
            steps {
                echo 'BUILD EXECUTION STARTED'
                sh 'go version'
                sh 'go get ./...'
                sh 'docker build . -t homieclips/hls-converter'
            }
        }
        stage('deliver') {
            agent any
            steps {
                withCredentials([usernamePassword(credentialsId: 'dockerhub', passwordVariable: 'dockerhubPassword', usernameVariable: 'dockerhubUser')]) {
                sh "docker login -u ${env.dockerhubUser} -p ${env.dockerhubPassword}"
                sh 'docker push homieclips/hls-converter'
                }
            }
        }
    }
}