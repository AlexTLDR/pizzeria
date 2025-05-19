# Docker Volume Management for Pizzeria

This document explains how to manage persistent data volumes for the Pizzeria application.

## Overview

The Pizzeria application uses Docker named volumes to store persistent data:

- `pizzeria_db_data`: Contains the SQLite database with all menu items and application data
- `pizzeria_images`: Contains uploaded menu item images

Using named volumes instead of bind mounts ensures data persistence across container rebuilds and restarts.

## Viewing Volumes

List all Docker volumes:

```bash
docker volume ls
```

Inspect a specific volume:

```bash
docker volume inspect pizzeria_db_data
```

## Backing Up Volumes

Regular backups are essential. Here's how to backup your volumes:

### 1. Database Backup

```bash
# Create a temporary container to access the volume
docker run --rm -v pizzeria_db_data:/source -v $(pwd):/backup alpine \
  tar czvf /backup/pizzeria_db_backup_$(date +%Y%m%d).tar.gz /source
```

### 2. Images Backup

```bash
# Create a temporary container to access the volume
docker run --rm -v pizzeria_images:/source -v $(pwd):/backup alpine \
  tar czvf /backup/pizzeria_images_backup_$(date +%Y%m%d).tar.gz /source
```

## Restoring Volumes

If you need to restore data from backups:

### 1. Database Restore

```bash
# First, stop the running containers
docker compose down

# Create a new volume if it doesn't exist
docker volume create pizzeria_db_data

# Restore from backup
docker run --rm -v pizzeria_db_data:/target -v $(pwd):/backup alpine \
  sh -c "rm -rf /target/* && tar xzvf /backup/pizzeria_db_backup_YYYYMMDD.tar.gz -C /target --strip-components=1"
```

### 2. Images Restore

```bash
# Create a new volume if it doesn't exist
docker volume create pizzeria_images

# Restore from backup
docker run --rm -v pizzeria_images:/target -v $(pwd):/backup alpine \
  sh -c "rm -rf /target/* && tar xzvf /backup/pizzeria_images_backup_YYYYMMDD.tar.gz -C /target --strip-components=1"

# Restart the application
docker compose up -d
```

## Environment-Specific Considerations

### Development

- In development, you might prefer using bind mounts for easier file editing:
  ```yaml
  volumes:
    - ./db:/app/db
    - ./static/images/menu:/app/static/images/menu
  ```

### Production (GCP)

- Named volumes work well on a single VM
- For multi-VM deployments, consider:
  - Google Cloud SQL for the database
  - Google Cloud Storage for images
  - See section below on migrating to cloud services

## Troubleshooting

### Volume permissions issues

If you encounter permission errors:

```bash
# Reset permissions in the volume
docker run --rm -v pizzeria_db_data:/data alpine chmod -R 777 /data
```

### Missing data after deployment

Check if volumes were created properly:

```bash
docker volume ls
docker compose down
docker compose up -d
```

### Corrupt database

If the SQLite database becomes corrupt:

1. Stop the application: `docker compose down`
2. Restore from the most recent backup using the instructions above
3. If no backup exists, you may need to recreate the volume: 
   ```bash
   docker volume rm pizzeria_db_data
   docker volume create pizzeria_db_data
   ```

## Future Considerations: Migrating to Cloud Services

For increased scalability and reliability, consider:

1. **Migrating to Google Cloud SQL**:
   - Create a Cloud SQL instance (PostgreSQL or MySQL)
   - Update application code to use the new database
   - Export data from SQLite and import to Cloud SQL

2. **Using Google Cloud Storage for images**:
   - Create a GCS bucket
   - Update application code to use the GCS API for image storage
   - Upload existing images to the bucket