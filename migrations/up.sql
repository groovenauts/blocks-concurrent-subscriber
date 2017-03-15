CREATE TABLE `pipeline_jobs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `pipeline` varchar(255) NOT NULL,
  `job_message_id` varchar(255) DEFAULT NULL,
  `progress` int(11) NOT NULL,
  PRIMARY KEY (`id`)
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
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
