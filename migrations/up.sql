CREATE TABLE `pipeline_jobs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `pipeline` varchar(255) NOT NULL,
  `job_message_id` varchar(255) DEFAULT NULL,
  `progress` int(11) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `index_pipeline_jobs_on_job_message_id` (`job_message_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `pipeline_job_logs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `pipeline` varchar(255) NOT NULL,
  `job_message_id` varchar(255) NOT NULL,
  `publish_time` datetime NOT NULL,
  `progress` int(11) NOT NULL,
  `completed` tinyint NOT NULL,
  `log_level` varchar(10) NOT NULL,
  `log_message` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `index_pipeline_job_logs_on_publish_time` (`publish_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
