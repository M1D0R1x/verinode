-- Migration down: 000003_index_and_surveillance.down.sql

DROP TABLE IF EXISTS surveillance_flags CASCADE;
DROP TABLE IF EXISTS index_contributions CASCADE;
DROP TABLE IF EXISTS index_observations CASCADE;
DROP TABLE IF EXISTS index_series CASCADE;
