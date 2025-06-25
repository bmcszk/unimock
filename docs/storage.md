# Internal Storage Architecture

This document describes the internal storage implementation of Unimock. This information is for developers and contributors who need to understand the internal architecture.

## Architecture Overview

Unimock uses an in-memory storage system with thread-safe operations to store and retrieve mock data. This implementation is completely internal and not exposed to users directly.

## Internal Data Structures

The storage system maintains several internal data structures:

1. **Data Map** - Maps internal storage IDs to MockData objects
2. **ID Map** - Maps external IDs to internal storage IDs
3. **Path Map** - Maps paths to lists of storage IDs

This architecture allows for efficient lookup by either ID or path while maintaining thread-safety for concurrent operations.

## MockData Structure

Each stored item contains:

- **Path** - The original request path (without trailing slashes)
- **Location** - The full resource path with ID
- **ContentType** - The content type of the stored data
- **Body** - The raw data bytes

## Internal Operations

### Create

Creates a new resource with one or more external IDs:

- Generates a unique internal storage ID (UUID)
- Maps all external IDs to the internal storage ID
- Maps the path to the storage ID
- Stores the data with content type and metadata
- Returns conflict error if any of the IDs already exist

### Get

Retrieves a resource by ID:

- Looks up the internal storage ID from the external ID
- Returns the stored data with its original content type
- Returns not found error if ID doesn't exist

### GetByPath

Retrieves all resources stored at a specific path:

- Returns all items mapped to the given path
- Performs prefix matching for deep paths
- Returns empty array if no items found

### Update

Updates an existing resource:

- Replaces the data for the given ID
- Updates path mappings if path has changed
- Returns not found error if ID doesn't exist

### Delete

Deletes a resource by ID:

- Removes all external ID mappings to the storage ID
- Removes path mappings
- Removes the data from storage
- Returns not found error if ID doesn't exist

## Thread Safety Implementation

All storage operations use mutex locks to ensure thread-safety:

- Read operations use a read lock (RLock)
- Write operations use a write lock (Lock)
- This implementation allows concurrent reads but serializes writes

## Path Handling Details

- All paths are stored without trailing slashes
- Location field stores the full resource path with ID
- Path map supports prefix matching for deep paths
- Example paths:
  - `/users` - Collection path
  - `/users/123` - Resource path with ID "123"
  - `/users/123/orders/456` - Deep path with ID "456" 
