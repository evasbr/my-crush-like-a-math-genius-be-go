-- SQL Seed down-migration: seed_roles
DELETE FROM roles WHERE name IN ('SUPER_ADMIN', 'SUB_ADMIN');
