DELETE FROM temp_imports WHERE created_at < NOW() - INTERVAL '7 days';
