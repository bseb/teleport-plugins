resource "teleport_database" "test" {
    metadata = {
        name    = "test"
        expires = "2032-10-12T07:20:50Z"
        labels  = {					
            example = "yes"
            "teleport.dev/origin" = "dynamic"
        }
    }

    spec = {
        protocol = "postgres"
        uri = "localhost"
    }
}
