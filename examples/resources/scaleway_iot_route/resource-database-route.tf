### Database Route

resource "scaleway_iot_route" "main" {
  name   = "default"
  hub_id = scaleway_iot_hub.main.id
  topic  = "#"
  database {
    query    = <<-EOT
			INSERT INTO measurements(
				push_time,
				report_time,
				station_id,
				temperature,
				humidity
			) VALUES (
				NOW(),
				TIMESTAMP 'epoch' + (($PAYLOAD::jsonb->'last_reported')::integer * INTERVAL '1 second'),
				($PAYLOAD::jsonb->'station_id')::uuid,
				($PAYLOAD::jsonb->'temperature')::decimal,
				($PAYLOAD::jsonb->'humidity'):decimal:
			);
			EOT
    host     = scaleway_rdb_instance.iot.endpoint_ip
    port     = scaleway_rdb_instance.iot.endpoint_port
    dbname   = "rdb"
    username = scaleway_rdb_instance.iot.user_name
    password = scaleway_rdb_instance.iot.password
  }
}

resource "scaleway_iot_hub" "main" {
  name         = "main"
  product_plan = "plan_shared"
}

resource "scaleway_rdb_instance" "iot" {
  name      = "iot"
  node_type = "db-dev-s"
  engine    = "PostgreSQL-12"
  user_name = "root"
  password  = "T3stP4ssw0rdD0N0tUs3!"
}
