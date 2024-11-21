job "httpecho" {
    datacenters = ["dc1"]

    group "echo" {
        count = 5
        network {
            port "http" {}
        }

        task "server" {
            driver = "docker"

            config {
                image = "hashicorp/http-echo:latest"

                ports = ["http"]

                args = [
                    "-listen", ":${NOMAD_PORT_http}",
                    "-text", "Hello and welcome to ${NOMAD_IP_http} running on port ${NOMAD_PORT_http}",
                ]
            }
        }
    }
}