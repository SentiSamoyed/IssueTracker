pipeline {
    agent any

    stages {
        stage('Deploy') {
            environment {
                NJU_PASSWORD = credentials('NJU_PASSWORD')
                GH_TOKEN = credentials('GH_TOKEN')
            }
            when {
                anyOf {
                    branch 'master'
                }
            }
            steps {
                echo 'Deploying....'
                sh 'sudo docker compose down'
                sh 'sudo docker compose up -d --build'
            }
        }
    }
}