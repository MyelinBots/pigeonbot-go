-- Migration to lowercase all player names and merge duplicates
-- This handles cases where "JohnDoe" and "johndoe" exist as separate rows

-- Step 1: Create a temp table with merged data
CREATE TEMP TABLE player_merged AS
SELECT
    MIN(id) as id,
    LOWER(name) as name,
    network,
    channel,
    SUM(points) as points,
    SUM(count) as count,
    SUM(eggs) as eggs,
    SUM(rare_eggs) as rare_eggs,
    MIN(created_at) as created_at,
    MAX(updated_at) as updated_at,
    NULL::timestamp as deleted_at
FROM player
WHERE deleted_at IS NULL
GROUP BY LOWER(name), network, channel;

-- Step 2: Delete all existing rows (we'll re-insert merged ones)
DELETE FROM player WHERE deleted_at IS NULL;

-- Step 3: Insert merged data back
INSERT INTO player (id, name, network, channel, points, count, eggs, rare_eggs, created_at, updated_at, deleted_at)
SELECT id, name, network, channel, points, count, eggs, rare_eggs, created_at, updated_at, deleted_at
FROM player_merged;

-- Step 4: Clean up temp table
DROP TABLE player_merged;
