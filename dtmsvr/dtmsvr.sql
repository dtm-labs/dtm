-- CREATE DATABASE `dtm` /*!40100 DEFAULT CHARACTER SET utf8mb4 */;

-- use dtm;

drop table IF EXISTS saga;
CREATE TABLE `saga` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `gid` varchar(45) NOT NULL COMMENT '事务全局id',
  `steps` json NOT NULL COMMENT 'saga的所有步骤',
  `status` varchar(45) NOT NULL COMMENT '全局事务的状态 prepared | processing | finished | rollbacked',
  `trans_query` varchar(128) NOT NULL COMMENT '事务未决状态的查询api',
  `finish_time` datetime DEFAULT NULL,
  `rollback_time` datetime DEFAULT NULL,
  `create_time` datetime DEFAULT NULL,
  `update_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `gid` (`gid`),
  KEY `create_time` (`create_time`),
  KEY `update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

drop table IF EXISTS saga_step;
CREATE TABLE `saga_step` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `gid` varchar(45) NOT NULL COMMENT '事务全局id',
  `data` json DEFAULT NULL COMMENT '请求所携带的数据',
  `step` int(11) NOT NULL COMMENT '处于saga中的第几步',
  `url` varchar(128) NOT NULL COMMENT '动作关联的url',
  `type` varchar(45) NOT NULL COMMENT 'saga的所有步骤',
  `status` varchar(45) NOT NULL COMMENT '步骤的状态 prepared | finished | rollbacked',
  `finish_time` datetime DEFAULT NULL,
  `rollback_time` datetime DEFAULT NULL,
  `create_time` datetime DEFAULT NULL,
  `update_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `gid` (`gid`,`step`),
  KEY `create_time` (`create_time`),
  KEY `update_time` (`update_time`)
) ENGINE=InnoDB AUTO_INCREMENT=17 DEFAULT CHARSET=utf8mb4;

drop table IF EXISTS trans_log;
CREATE TABLE `trans_log` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `gid` varchar(45) NOT NULL COMMENT '事务全局id',
  `step` int(11) DEFAULT NULL COMMENT 'saga的步骤',
  `action` varchar(45) DEFAULT NULL COMMENT '行为',
  `status` varchar(45) NOT NULL COMMENT '状态',
  `detail` json NOT NULL COMMENT '行为之后的status',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `gid` (`gid`),
  KEY `create_time` (`create_time`)
) ENGINE=InnoDB AUTO_INCREMENT=48 DEFAULT CHARSET=utf8mb4;

