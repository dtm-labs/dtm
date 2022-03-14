CREATE DATABASE IF NOT EXISTS dtm
/*!40100 DEFAULT CHARACTER SET utf8mb4 */
;
drop table IF EXISTS dtm.trans_global;
CREATE TABLE if not EXISTS dtm.trans_global (
  `id` bigint(22) NOT NULL AUTO_INCREMENT,
  `gid` varchar(128) NOT NULL COMMENT 'global transaction id',
  `trans_type` varchar(45) not null COMMENT 'transaction type: saga | xa | tcc | msg',
  `status` varchar(12) NOT NULL COMMENT 'tranaction status: prepared | submitted | aborting | finished | rollbacked',
  `query_prepared` varchar(128) NOT NULL COMMENT 'url to check for 2-phase message',
  `protocol` varchar(45) not null comment 'protocol: http | grpc | json-rpc',
  `create_time` datetime DEFAULT NULL,
  `update_time` datetime DEFAULT NULL,
  `finish_time` datetime DEFAULT NULL,
  `rollback_time` datetime DEFAULT NULL,
  `options` varchar(1024) DEFAULT 'options for transaction like: TimeoutToFail, RequestTimeout',
  `custom_data` varchar(256) DEFAULT '' COMMENT 'custom data for transaction',
  `next_cron_interval` int(11) default null comment 'next cron interval. for use of cron job',
  `next_cron_time` datetime default null comment 'next time to process this trans. for use of cron job',
  `owner` varchar(128) not null default '' comment 'who is locking this trans',
  `ext_data` TEXT comment 'extended data for this trans',
  PRIMARY KEY (`id`,`gid`),
  UNIQUE KEY `id` (`id`,`gid`),
  UNIQUE KEY `gid` (`gid`),
  key `owner`(`owner`),
  key `status_next_cron_time` (`status`, `next_cron_time`) comment 'cron job will use this index to query trans'
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 shardkey=gid;
drop table IF EXISTS dtm.trans_branch_op;
CREATE TABLE IF NOT EXISTS dtm.trans_branch_op (
  `id` bigint(22) NOT NULL AUTO_INCREMENT,
  `gid` varchar(128) NOT NULL COMMENT 'global transaction id',
  `url` varchar(128) NOT NULL COMMENT 'the url of this op',
  `data` TEXT COMMENT 'request body, depreceated',
  `bin_data` BLOB COMMENT 'request body',
  `branch_id` VARCHAR(128) NOT NULL COMMENT 'transaction branch ID',
  `op` varchar(45) NOT NULL COMMENT 'transaction operation type like: action | compensate | try | confirm | cancel',
  `status` varchar(45) NOT NULL COMMENT 'transaction op status: prepared | succeed | failed',
  `finish_time` datetime DEFAULT NULL,
  `rollback_time` datetime DEFAULT NULL,
  `create_time` datetime DEFAULT NULL,
  `update_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`,`gid`),
  UNIQUE KEY `id` (`id`,`gid`),
  UNIQUE KEY `gid_uniq` (`gid`, `branch_id`, `op`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 shardkey=gid;
