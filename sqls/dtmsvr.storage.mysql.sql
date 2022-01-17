CREATE DATABASE IF NOT EXISTS dtm
/*!40100 DEFAULT CHARACTER SET utf8mb4 */
;
drop table IF EXISTS dtm.trans_global;
CREATE TABLE if not EXISTS dtm.trans_global (
  `id` bigint(22) NOT NULL AUTO_INCREMENT,
  `gid` varchar(128) NOT NULL COMMENT '事务全局id',
  `trans_type` varchar(45) not null COMMENT '事务类型: saga | xa | tcc | msg',
  -- `data` TEXT COMMENT '事务携带的数据', -- 影响性能，不必要存储
  `status` varchar(12) NOT NULL COMMENT '全局事务的状态 prepared | submitted | aborting | finished | rollbacked',
  `query_prepared` varchar(128) NOT NULL COMMENT 'prepared状态事务的查询api',
  `protocol` varchar(45) not null comment '通信协议 http | grpc',
  `create_time` datetime DEFAULT NULL,
  `update_time` datetime DEFAULT NULL,
  `finish_time` datetime DEFAULT NULL,
  `rollback_time` datetime DEFAULT NULL,
  `options` varchar(1024) DEFAULT '',
  `custom_data` varchar(256) DEFAULT '',
  `next_cron_interval` int(11) default null comment '下次定时处理的间隔',
  `next_cron_time` datetime default null comment '下次定时处理的时间',
  `owner` varchar(128) not null default '' comment '正在处理全局事务的锁定者',
  `ext_data` TEXT comment 'global扩展字段的数据',
  PRIMARY KEY (`id`),
  UNIQUE KEY `gid` (`gid`),
  key `owner`(`owner`),
  key `status_next_cron_time` (`status`, `next_cron_time`) comment '这个索引用于查询超时的全局事务，能够合理的走索引'
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
drop table IF EXISTS dtm.trans_branch_op;
CREATE TABLE IF NOT EXISTS dtm.trans_branch_op (
  `id` bigint(22) NOT NULL AUTO_INCREMENT,
  `gid` varchar(128) NOT NULL COMMENT '事务全局id',
  `url` varchar(128) NOT NULL COMMENT '动作关联的url',
  `data` TEXT COMMENT '请求所携带的数据',
  `bin_data` BLOB COMMENT 'grpc的二进制数据',
  `branch_id` VARCHAR(128) NOT NULL COMMENT '事务分支ID',
  `op` varchar(45) NOT NULL COMMENT '事务分支类型 saga_action | saga_compensate | xa',
  `status` varchar(45) NOT NULL COMMENT '步骤的状态 submitted | finished | rollbacked',
  `finish_time` datetime DEFAULT NULL,
  `rollback_time` datetime DEFAULT NULL,
  `create_time` datetime DEFAULT NULL,
  `update_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `gid_uniq` (`gid`, `branch_id`, `op`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
