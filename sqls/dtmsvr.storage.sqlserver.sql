if db_id('dtm') is null
begin 
	CREATE DATABASE dtm
end;

drop table IF EXISTS dtm.dbo.trans_global;
CREATE TABLE dtm.dbo.trans_global (
  id bigint NOT NULL IDENTITY,
  gid varchar(128) NOT NULL , -- COMMENT 'global transaction id',
  trans_type varchar(45) not null , -- COMMENT 'transaction type: saga | xa | tcc | msg',
  status varchar(12) NOT NULL , -- COMMENT 'transaction status: prepared | submitted | aborting | succeed | failed',
  query_prepared varchar(1024) NOT NULL , -- COMMENT 'url to check for msg|workflow',
  protocol varchar(45) not null , -- COMMENT 'protocol: http | grpc | json-rpc',
  create_time datetimeoffset DEFAULT NULL,
  update_time datetimeoffset DEFAULT NULL,
  finish_time datetimeoffset DEFAULT NULL,
  rollback_time datetimeoffset DEFAULT NULL,
  options varchar(1024) DEFAULT '' , -- COMMENT 'options for transaction like: TimeoutToFail, RequestTimeout',
  custom_data varchar(1024) DEFAULT '' , -- COMMENT 'custom data for transaction',
  next_cron_interval int default null , -- COMMENT 'next cron interval. for use of cron job',
  next_cron_time datetimeoffset default null , -- COMMENT 'next time to process this trans. for use of cron job',
  owner varchar(128) not null default '' , -- COMMENT 'who is locking this trans',
  ext_data VARCHAR(max) , -- COMMENT 'extra data for this trans. currently used in workflow pattern',
  result varchar(1024) DEFAULT '' , -- COMMENT 'result for transaction',
  rollback_reason varchar(1024) DEFAULT '' , -- COMMENT 'rollback reason for transaction',
  PRIMARY KEY (id),
  CONSTRAINT gid UNIQUE (gid) WITH(IGNORE_DUP_KEY = ON)
);
CREATE INDEX[owner] ON [dtm].[dbo].[trans_global]([owner] ASC)
CREATE INDEX[status_next_cron_time] ON [dtm].[dbo].[trans_global]([status] ASC, next_cron_time ASC) ---- COMMENT 'cron job will use this index to query trans'

drop table IF EXISTS dtm.dbo.trans_branch_op;
CREATE TABLE dtm.dbo.trans_branch_op (
  id bigint NOT NULL IDENTITY,
  gid varchar(128) NOT NULL , -- COMMENT 'global transaction id',
  url varchar(1024) NOT NULL , -- COMMENT 'the url of this op',
  data VARCHAR(max) , -- COMMENT 'request body, depreceated',
  bin_data VARBINARY(max) , -- COMMENT 'request body',
  branch_id VARCHAR(128) NOT NULL , -- COMMENT 'transaction branch ID',
  op varchar(45) NOT NULL , -- COMMENT 'transaction operation type like: action | compensate | try | confirm | cancel',
  status varchar(45) NOT NULL , -- COMMENT 'transaction op status: prepared | succeed | failed',
  finish_time datetimeoffset DEFAULT NULL,
  rollback_time datetimeoffset DEFAULT NULL,
  create_time datetimeoffset DEFAULT NULL,
  update_time datetimeoffset DEFAULT NULL,
  PRIMARY KEY (id),
  CONSTRAINT gid_uniq UNIQUE  (gid, branch_id, op)
);
drop table IF EXISTS dtm.dbo.kv;
CREATE TABLE dtm.dbo.kv (
  id bigint NOT NULL IDENTITY,
  cat varchar(45) NOT NULL , -- COMMENT 'the category of this data',
  k varchar(128) NOT NULL,
  v VARCHAR(max),
  version bigint default 1 , -- COMMENT 'version of the value',
  create_time datetimeoffset default NULL,
  update_time datetimeoffset DEFAULT NULL,
  PRIMARY KEY (id),
  CONSTRAINT uniq_k UNIQUE (cat, k)
);
