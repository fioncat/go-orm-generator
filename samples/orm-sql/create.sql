-- ------------------------------------------------------------
-- go-gendb v0.3
--
-- https://github.com/fioncat/go-gendb
--
-- This file is auto generated. DO NOT EDIT.
-- ------------------------------------------------------------
USE test;

CREATE TABLE `user` (

  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '用户自增id',
  `name` varchar(256) NOT NULL DEFAULT '' COMMENT '用户名称',
  `phone` varchar(11) NOT NULL DEFAULT '' COMMENT '用户电话号码',
  `code` varchar(256) DEFAULT 'AB123' COMMENT '用户的唯一编码',
  `is_removed` tinyint(2) DEFAULT NULL COMMENT '用户是否被删除',
  `create_date` bigint DEFAULT NULL COMMENT '用户创建时间',

  PRIMARY KEY (`id`),

  UNIQUE INDEX `unique_user_Code`(`code`),

  INDEX `index_user_Name_Phone`(`name`,`phone`),
  INDEX `index_user_CreateDate`(`create_date`),
  INDEX `index_user_Name`(`name`),
  INDEX `index_user_Phone`(`phone`)

) ENGINE=InnoDB COMMENT '用户表';