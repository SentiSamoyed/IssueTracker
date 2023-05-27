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
                sh 'sudo http_proxy=http://127.0.0.1:7890 https_proxy=http://127.0.0.1:7890 bash ./build.sh'
                sh 'sudo docker compose down'
                sh 'sudo NJU_PASSWORD=$NJU_PASSWORD GH_TOKEN=$GH_TOKEN docker compose up -d --build'
            }
        }
    }
}