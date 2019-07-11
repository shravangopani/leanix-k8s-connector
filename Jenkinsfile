pipeline {
    agent any

    stages {
        stage('Test') {
            steps {
                sh 'make test'
            }
        }
        stage('Build') {
            when { 
                anyOf { 
                    branch 'master'
                    branch 'develop' 
                } 
            }
            steps {
                sh 'make'
                sh 'make image'
                sh 'make push'
            }
        }
        stage('Deploy to Test') {
            when { 
                anyOf { 
                    branch 'master'
                    branch 'develop' 
                } 
            }
            steps {
                echo 'Here we need to run helm command to deploy to the leanix int cluster'
            }
        }
        stage('Release approval'){
            when {
                branch 'master'
            }
            input "Release new version?"
        }
        stage('Release') {
            when {
                branch 'master'
            }
            steps {
                echo 'Set the version variable as default for image tag in helm chart'
            }
        }
    }
}