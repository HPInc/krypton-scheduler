CREATE KEYSPACE IF NOT EXISTS scheduler 
  WITH replication = {
    'class': 'SimpleStrategy', 
    'replication_factor': '1'
  };