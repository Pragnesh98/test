CREATE TABLE `callbacks` (
  `id` int NOT NULL AUTO_INCREMENT,
  `sid` varchar(40) NOT NULL DEFAULT '',
  `created_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00',
  `updated_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00',
  `callback_url` varchar(200) DEFAULT NULL,
  `status` varchar(20) NOT NULL DEFAULT '',
  `payload` json DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=199 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC
