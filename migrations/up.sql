CREATE TABLE `pipeline_jobs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `pipeline` varchar(255) NOT NULL,
  `message_id` varchar(255) DEFAULT NULL,
  `status` int(11) NOT NULL,
  PRIMARY KEY (`id`),
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `pipeline_job_logs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `pipeline` varchar(255) NOT NULL,
  `message_id` varchar(255) NOT NULL,
  `status` int(11) NOT NULL,
  `publish_time` datetime NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
