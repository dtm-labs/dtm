CREATE SCHEMA if not EXISTS dtm
/* SQLINES DEMO *** RACTER SET utf8mb4 */
;
drop table IF EXISTS dtm.trans_global;
-- SQLINES LICENSE FOR EVALUATION USE ONLY
CREATE SEQUENCE if not EXISTS dtm.trans_global_seq;
CREATE TABLE if not EXISTS dtm.trans_global (
  id bigint NOT NULL DEFAULT NEXTVAL ('dtm.trans_global_seq'),
  gid varchar(128) NOT NULL,
  trans_type varchar(45) not null,
  status varchar(45) NOT NULL,
  query_prepared varchar(128) NOT NULL,
  protocol varchar(45) not null,
  create_time timestamp(0) DEFAULT NULL,
  update_time timestamp(0) DEFAULT NULL,
  commit_time timestamp(0) DEFAULT NULL,
  finish_time timestamp(0) DEFAULT NULL,
  rollback_time timestamp(0) DEFAULT NULL,
  options varchar(256) DEFAULT '',
  custom_data varchar(256) DEFAULT '',
  next_cron_interval int default null,
  next_cron_time timestamp(0) default null,
  owner varchar(128) not null default '',
  PRIMARY KEY (id),
  CONSTRAINT gid UNIQUE (gid)
);
create index if not EXISTS owner on dtm.trans_global(owner);
CREATE INDEX if not EXISTS create_time ON dtm.trans_global (create_time);
CREATE INDEX if not EXISTS update_time ON dtm.trans_global (update_time);
create index if not EXISTS status_next_cron_time on dtm.trans_global (status, next_cron_time);
drop table IF EXISTS dtm.trans_branch;
-- SQLINES LICENSE FOR EVALUATION USE ONLY
CREATE SEQUENCE if not EXISTS dtm.trans_branch_seq;
CREATE TABLE IF NOT EXISTS dtm.trans_branch (
  id bigint NOT NULL DEFAULT NEXTVAL ('dtm.trans_branch_seq'),
  gid varchar(128) NOT NULL,
  url varchar(128) NOT NULL,
  data TEXT,
  branch_id VARCHAR(128) NOT NULL,
  branch_type varchar(45) NOT NULL,
  status varchar(45) NOT NULL,
  finish_time timestamp(0) DEFAULT NULL,
  rollback_time timestamp(0) DEFAULT NULL,
  create_time timestamp(0) DEFAULT NULL,
  update_time timestamp(0) DEFAULT NULL,
  PRIMARY KEY (id),
  CONSTRAINT gid_uniq UNIQUE (gid, branch_id, branch_type)
);
CREATE INDEX if not EXISTS create_time ON dtm.trans_branch (create_time);
CREATE INDEX if not EXISTS update_time ON dtm.trans_branch (update_time);
drop table IF EXISTS dtm.trans_log;
-- SQLINES LICENSE FOR EVALUATION USE ONLY
CREATE SEQUENCE if not EXISTS dtm.trans_log_seq;
CREATE TABLE IF NOT EXISTS dtm.trans_log (
  id bigint NOT NULL DEFAULT NEXTVAL ('dtm.trans_log_seq'),
  gid varchar(128) NOT NULL,
  branch_id varchar(128) DEFAULT NULL,
  action varchar(45) DEFAULT NULL,
  old_status varchar(45) NOT NULL DEFAULT '',
  new_status varchar(45) NOT NULL,
  detail TEXT NOT NULL,
  create_time timestamp(0) DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
);
CREATE INDEX if not EXISTS gid ON dtm.trans_log (gid);
CREATE INDEX if not EXISTS create_time ON dtm.trans_log (create_time);