DELETE FROM pipeline_jobs;
INSERT INTO pipeline_jobs
  (pipeline, job_message_id, progress, created_at, updated_at)
  VALUES
  ('pipeline01', 'jm01', 1, now(), now()),
  ('pipeline01', 'jm02', 2, now(), now()),
  ('pipeline01', 'jm03', 3, now(), now()),
  ('pipeline01', 'jm04', 4, now(), now()),
  ('pipeline01', 'jm05', 5, now(), now()),
  ('pipeline02', 'jm01', 1, now(), now()),
  ('pipeline02', 'jm02', 2, now(), now()),
  ('pipeline02', 'jm03', 3, now(), now()),
  ('pipeline02', 'jm04', 4, now(), now()),
  ('pipeline02', 'jm05', 5, now(), now());

DELETE FROM pipeline_job_logs;
