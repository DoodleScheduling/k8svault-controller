import java.util.regex.Pattern
import org.jenkinsci.plugins.pipeline.modeldefinition.Utils

node("jenkins") {
  ansiColor("xterm") {
    stage('checkout') {
      checkout(scm)
      dockerAuth()
    }

    stage("build") {
      sh 'make test docker-build'
      sh 'helm3 lint chart/k8svault-controller'
    }

    stage("publish") {
      if (!env.TAG_NAME) {
        echo "skip packaging for no tagged release"
      } else {
        def (_,major,minor,patch,group,label,build) = (env.TAG_NAME =~ /^v(\d{1,3})\.(\d{1,3})\.(\d{1,3})(?:(-([A-Za-z0-9]+)))?$/)[0]

        if (!major && !minor && !patch) {
          throw new Exception("Invalid tag detected, requires semantic version")
        }

        version = "$major.$minor.$patch$group"
        sh 'make docker-build IMG=nexus.doodle.com:5000/devops/k8svault-controller:v$version'
        sh 'make test docker-push'

        bumpChartVersion("v${version}")
        bumpImageVersion(version)

        tgz="k8ssecret-controller-${version}.tgz"
        sh "helm3 package chart/k8svault-controller"

        if (label) {
          publish(tgz, "nexus-staging")
        } else {
          publish(tgz, "nexus-staging")
          publish(tgz, "nexus-production")
        }
      }
    }
  }
}

void dockerAuth() {
  // nexus repository
  withCredentials([[
                       $class          : 'UsernamePasswordMultiBinding',
                       credentialsId   : 'nexus',
                       usernameVariable: 'NEXUS_USER',
                       passwordVariable: 'NEXUS_PASSWORD'
                   ]]) {
    sh "docker login nexus.doodle.com:5000 -u ${env.NEXUS_USER} -p ${env.NEXUS_PASSWORD}"
  }

  // docker hub
  withCredentials([[
                       $class          : 'UsernamePasswordMultiBinding',
                       credentialsId   : 'dockerhub',
                       usernameVariable: 'DOCKERHUB_USER',
                       passwordVariable: 'DOCKERHUB_PASSWORD'
                   ]]) {
    sh "docker login -u ${env.DOCKERHUB_USER} -p ${env.DOCKERHUB_PASSWORD}"
  }
}

def bumpImageVersion(String version) {
  echo "Update image tag"
  def valuesFile = "./chart/k8svault-controller/values.yaml"
  def valuesData = readYaml file: valuesFile
  chartData.image.tag = version

  sh "rm $valuesFile"
  writeYaml file: valuesFile, data: valuesData
}

def bumpChartVersion(String version) {
  // Bump chart version
  echo "Update chart version"
  def chartFile = "./chart/k8svault-controller/Chart.yaml"
  def chartData = readYaml file: chartFile
  chartData.version = version
  chartData.appVersion = version

  sh "rm $chartFile"
  writeYaml file: chartFile, data: chartData
}

def publish(String tgz, String repository) {
  echo "Push chart ${tgz} to helm repository ${repository}"

  withCredentials([[
    $class          : 'UsernamePasswordMultiBinding',
    credentialsId   : 'nexus',
    usernameVariable: 'NEXUS_USER',
    passwordVariable: 'NEXUS_PASSWORD'
  ]]) {
    sh "curl -u \"${env.NEXUS_USER}:${env.NEXUS_PASSWORD}\" https://nexus.doodle.com/repository/${repository}/ --upload-file $tgz --fail"
  }
}