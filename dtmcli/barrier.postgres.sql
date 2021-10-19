create schema if not exists dtm_barrier;

drop table if exists dtm_barrier.barrier;

CREATE SEQUENCE if not EXISTS dtm_barrier.barrier_seq;

create table if not exists dtm_barrier.barrier(
  id int NOT NULL DEFAULT NEXTVAL ('dtm_barrier.barrier_seq'),
  trans_type varchar(45) default '' ,
  gid varchar(128) default'',
  branch_id varchar(128) default '',
  branch_type varchar(45) default '',
  barrier_id varchar(45) default '',
  reason varchar(45) default '',
  create_time timestamp(0) DEFAULT NULL,
  update_time timestamp(0) DEFAULT NULL,
  PRIMARY KEY(id),
  CONSTRAINT uniq_barrier unique(gid, branch_id, branch_type, barrier_id)
);

