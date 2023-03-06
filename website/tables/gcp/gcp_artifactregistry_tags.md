# Table: gcp_artifactregistry_tags

https://cloud.google.com/artifact-registry/docs/reference/rest/v1/projects.locations.repositories.packages.tags#Tag

The composite primary key for this table is (**project_id**, **name**).

## Relations

This table depends on [gcp_artifactregistry_packages](gcp_artifactregistry_packages).

## Columns

| Name          | Type          |
| ------------- | ------------- |
|_cq_source_name|String|
|_cq_sync_time|Timestamp|
|_cq_id|UUID|
|_cq_parent_id|UUID|
|project_id (PK)|String|
|name (PK)|String|
|version|String|