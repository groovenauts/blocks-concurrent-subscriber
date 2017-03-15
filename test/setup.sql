DELETE FROM pipeline_jobs;
INSERT INTO pipeline_jobs
  (pipeline, job_message_id, progress)
  VALUES
  ('pipeline01', 'jm01', 1),
  ('pipeline01', 'jm02', 2),
  ('pipeline01', 'jm03', 3),
  ('pipeline01', 'jm04', 4),
  ('pipeline01', 'jm05', 5),
  ('pipeline02', 'jm01', 1),
  ('pipeline02', 'jm02', 2),
  ('pipeline02', 'jm03', 3),
  ('pipeline02', 'jm04', 4),
  ('pipeline02', 'jm05', 5);

DELETE FROM pipeline_job_logs;
