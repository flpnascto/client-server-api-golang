CREATE TABLE IF NOT EXISTS quotations (
  code_out VARCHAR(255) NOT NULL,
  code_in VARCHAR(255) NOT NULL,
  bid FLOAT NOT NULL,
  timestamp TIMESTAMP NOT NULL,
  PRIMARY KEY (code_out, code_in, timestamp)
);