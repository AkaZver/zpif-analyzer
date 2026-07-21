-- Extension for UUID generation (optional, not used in current schema)
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- The schema will be created automatically by GORM AutoMigrate
-- This file is here for future manual migrations if needed

-- Create initial admin user (password: admin)
-- This will be handled by the Go application seed function
