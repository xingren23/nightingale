set names utf8;
use n9e_uic;
ALTER TABLE team ADD COLUMN `nid` int unsigned NOT NULL COMMENT '关联服务树id';

use n9e_mon;
ALTER TABLE event_cur ADD COLUMN `alert_users` VARCHAR(1000) NOT NULL COMMENT '已报警用户';

use n9e_mon;
ALTER TABLE endpoint Add COLUMN `tags` VARCHAR(256) NULL COMMENT '标签';

use n9e_uic;
alter table team drop index ident;
CREATE UNIQUE INDEX u_ident_nid ON team (ident,nid);

use n9e_mon;

CREATE TABLE `config_info`
(
    `id`          BIGINT(20)    NOT NULL AUTO_INCREMENT,
    `cfg_group`   VARCHAR(50)   NOT NULL COMMENT '组',
    `cfg_key`     VARCHAR(100)  NOT NULL COMMENT '键',
    `cfg_value`   VARCHAR(1500) NOT NULL COMMENT '值',
    `create_by`   BIGINT(19)    NOT NULL DEFAULT '0' COMMENT '创建人',
    `create_time` datetime      NOT NULL COMMENT '创建时间',
    `update_by`   BIGINT(19)    NOT NULL DEFAULT '0' COMMENT '修改人',
    `update_time` datetime      NOT NULL COMMENT '修改时间',
    `status`      TINYINT(4)    NOT NULL COMMENT '状态(1启用，0停用，-1删除)',
    PRIMARY KEY (`id`),
    KEY `IDX_config_info` (`cfg_group`, `cfg_key`, `status`)
) ENGINE = INNODB
  DEFAULT CHARSET = utf8 COMMENT = '配置信息';

CREATE TABLE `monitor_item`
(
    `id`            BIGINT(20)   NOT NULL AUTO_INCREMENT COMMENT '唯一ID',
    `name`          VARCHAR(255) NOT NULL DEFAULT '' COMMENT '监控项名称',
    `metric`        VARCHAR(128) NOT NULL DEFAULT '' COMMENT '指标编码',
    `type`          VARCHAR(32)  NOT NULL DEFAULT 'GAUGE' COMMENT '指标类型',
    `category`      VARCHAR(32)  NOT NULL COMMENT '监控项类别',
    `endpoint_type` VARCHAR(50)  NOT NULL DEFAULT 'HOST' COMMENT 'Endpoint类型',
    `step`          INT(11)      NOT NULL DEFAULT '60' COMMENT '指标步长',
    `unit`          VARCHAR(20)           DEFAULT '' COMMENT '指标显示单位',
    `formula`       VARCHAR(100)          DEFAULT '' COMMENT '指标值转换公式',
    `re_formula`    VARCHAR(100)          DEFAULT '' COMMENT '指标值还原公式',
    `description`   VARCHAR(255)          DEFAULT '' COMMENT '描述信息',
    `influx_type`   VARCHAR(100)          DEFAULT '' COMMENT 'influxdb指标抽象类型',
    `measurement`   VARCHAR(255)          DEFAULT '' COMMENT 'influxdb表名称',
    `tags`          VARCHAR(500)          DEFAULT '' COMMENT '指标标签集合["",""]',
    `machine_type`  VARCHAR(20)  NOT NULL DEFAULT '' COMMENT '指标机器类型',
    `create_time`   datetime     NOT NULL COMMENT '创建时间',
    `create_by`     BIGINT(20)   NOT NULL DEFAULT '0' COMMENT '创建人',
    `update_time`   datetime     NOT NULL COMMENT '修改时间',
    `update_by`     BIGINT(20)   NOT NULL COMMENT '修改人',
    `status`        TINYINT(4)   NOT NULL DEFAULT '1' COMMENT '状态(1/启用，0/停用，-1/删除)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `UK_metric` (`metric`),
    KEY `index_category` (`category`),
    KEY `category_metric_status` (`category`, `metric`, `status`)
) ENGINE = INNODB
  DEFAULT CHARSET = utf8 COMMENT = '监控项元数据';
