@startuml
hide circle
hide methods
hide stereotypes
skinparam linetype ortho

entity "users" as users {
\*id : uuid <<PK>>
email : string <<unique>>
password_hash : string
name : string
}

entity "gardens" as gardens {
\*id : uuid <<PK>>
user_id : uuid <<FK>>
name : string
location_text : string
memo : string
}

entity "plants" as plants {
\*id : uuid <<PK>>
garden_id : uuid <<FK>>
name : string
variety : string
sowing_date : date
planting_date : date
expected_harvest_start_date : date
expected_harvest_end_date : date
memo : string
}

entity "care_logs" as care_logs {
\*id : uuid <<PK>>
plant_id : uuid <<FK>>
log_type : string
amount : numeric
memo : string
logged_at : timestamptz
}

entity "photos" as photos {
\*id : uuid <<PK>>
plant_id : uuid <<FK>>
s3_key : string
taken_at : timestamptz
}

users ||--o{ gardens : owns
gardens ||--o{ plants : contains
plants ||--o{ care_logs : has
plants ||--o{ photos : has
@enduml
