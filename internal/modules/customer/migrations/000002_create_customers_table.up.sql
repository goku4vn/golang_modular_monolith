-- Create customer status enum
DROP TYPE IF EXISTS "public"."customer_status";
CREATE TYPE "public"."customer_status" AS ENUM ('active', 'inactive', 'deleted');

-- Create customers table
CREATE TABLE "public"."customers" (
    "id" VARCHAR(36) NOT NULL PRIMARY KEY,
    "name" VARCHAR(255) NOT NULL,
    "email" VARCHAR(255) NOT NULL UNIQUE,
    "status" "public"."customer_status" NOT NULL DEFAULT 'active'::customer_status,
    "version" INTEGER NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX idx_customers_email ON "public"."customers" ("email");
CREATE INDEX idx_customers_status ON "public"."customers" ("status");
CREATE INDEX idx_customers_created_at ON "public"."customers" ("created_at");
CREATE INDEX idx_customers_name ON "public"."customers" ("name");

-- Create trigger to automatically update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_customers_updated_at
    BEFORE UPDATE ON "public"."customers"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
