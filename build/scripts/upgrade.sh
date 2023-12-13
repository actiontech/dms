#!/usr/bin/env bash
###
### upgrade.sh - migrate sqle database to dms database
###
### Usage:
###   upgrade.sh -edition=enterprise -s_host=10.186.62.50 -s_port=3306 -s_user=root -s_pwd=mysqlpass -d_host=10.186.62.50 -d_port=33063 -d_user=root -d_pwd=123
###
### Options:
###   -edition  community, enterprise
###   -s_host   connect to host in sqle
###   -s_port   port number to use for connection in sqle
###   -s_user   user for login in sqle
###   -s_pwd    password to use when connecting to server in sqle
###   -d_host   connect to host in dms
###   -d_port   port number to use for connection in dms
###   -d_user   user for login in dms
###   -d_pwd    password to use when connecting to server in dms
###   -h        help message

set -e

help() {
  sed -rn 's/^### ?//;T;p;' "$0"
}

if [[ $# == 0 ]] || [[ "$1" == "-h" ]]; then
  help
  exit 1
fi

TEMP=$(getopt -n "$0" -a -l "edition:,s_host:,s_port:,s_user:,s_pwd:,d_host:,d_port:,d_user:,d_pwd:" -- -- "$@")
# shellcheck disable=SC2181
[ $? -eq 0 ] || exit

eval set -- "$TEMP"

while [ $# -gt 0 ]
do
   case "$1" in
      --edition) edition="$2"; shift 2;;
      --s_host) s_host="$2"; shift 2;;
      --s_port) s_port="$2"; shift 2;;
      --s_user) s_user="$2"; shift 2;;
      --s_pwd) s_pwd="$2"; shift 2;;
      --d_host) d_host="$2"; shift 2;;
      --d_port) d_port="$2"; shift 2;;
      --d_user) d_user="$2"; shift 2;;
      --d_pwd) d_pwd="$2"; shift 2;;
      --) shift;;
   esac
done

tables=()
tables[0]="projects"
tables[1]="users"
tables[2]="user_groups"
tables[3]="user_group_users"
tables[4]="global_configuration_smtp"
tables[5]="global_configuration_wechat"
tables[6]="global_configuration_im"
tables[7]="global_configuration_webhook"
tables[8]="global_configuration_ldap"
tables[9]="global_configuration_oauth2"
tables[10]="global_configuration_personalise"
tables[11]="global_configuration_logo"
tables[12]="sync_instance_tasks"
tables[13]="rule_templates"
tables[14]="instances"
tables[15]="roles"
tables[16]="role_operations"
tables[17]="management_permissions"
tables[18]="project_user"
tables[19]="project_member_roles"
tables[20]="project_manager"
tables[21]="project_user_group"
tables[22]="project_member_group_roles"


database="sqleback"

# mysql -h "$d_host" -P "$d_port" -u"$d_user" -p"$d_pwd" -e "drop database if exists $database; create database $database;"

# 导入sql数据到备份数据库
# mysqldump -h "$s_host" -P "$s_port" -u"$s_user" -p"$s_pwd" sqle | mysql -h "$d_host" -P "$d_port" -u"$d_user" -p"$d_pwd" "$database";

funMigrateDMS(){
  mysql -h "$d_host" -P "$d_port" -u"$d_user" -p"$d_pwd" -v -e "
  # 切换数据库到dms
  use dms;

  # import projects
  truncate table projects;
  insert ignore into projects (uid, name, \`desc\`, create_user_uid, status, created_at, updated_at) select if(id=1, '700300', id), name, \`desc\`, if(create_user_id=1, '700200', create_user_id), status, created_at, updated_at from $database.projects where deleted_at is null;

  # import users
  insert ignore into users (uid, name, email, phone, wechat_id, password, user_authentication_type, stat, third_party_user_id, created_at, updated_at) select if(id=1, '700200', id), login_name, email, phone, wechat_id, password, if(user_authentication_type = '', 'dms', user_authentication_type), stat, third_party_user_id, created_at, updated_at from $database.users where deleted_at is null;

  # import user_groups
  insert ignore into user_groups (uid, name, description, stat, created_at, updated_at) select id, name, description, stat, created_at, updated_at from $database.user_groups where deleted_at is null;

  # import user_group_users
  insert ignore into user_group_users (user_group_uid, user_uid) select user_group_id, if(user_id=1, '700200', user_id) from $database.user_group_users;

  # import smtp_configurations
  insert ignore into smtp_configurations(uid, enable_smtp_notify, smtp_host, smtp_port, smtp_username, secret_smtp_password, is_skip_verify, created_at, updated_at) select id, enable_smtp_notify, smtp_host, smtp_port, smtp_username, secret_smtp_password, is_skip_verify, created_at, updated_at from  $database.global_configuration_smtp where deleted_at is null;

  # import we_chat_configurations
  insert ignore into we_chat_configurations (uid, enable_we_chat_notify, corp_id, encrypted_corp_secret, agent_id, safe_enabled, proxy_ip, created_at, updated_at) select id, enable_we_chat_notify, corp_id, encrypted_corp_secret, agent_id, safe_enabled, proxy_ip, created_at, updated_at from $database.global_configuration_wechat where deleted_at is null;

  # import im_configurations
  insert ignore into im_configurations (uid, app_key, app_secret, is_enable, process_code, type, created_at, updated_at) select id, app_key, encrypt_app_secret, is_enable, process_code, type, created_at, updated_at from $database.global_configuration_im where deleted_at is null;

  # import web_hook_configurations
  insert ignore into web_hook_configurations (uid, enable, max_retry_times, retry_interval_seconds, encrypted_token, url, created_at, updated_at) select id, enable, max_retry_times, retry_interval_seconds, encrypted_token, url, created_at, updated_at from $database.global_configuration_webhook where deleted_at is null;

  # import ldap_configurations
  insert ignore into ldap_configurations (uid, enable, enable_ssl, host, port, connect_dn, connect_secret_password, base_dn, user_name_rdn_key, user_email_rdn_key, created_at, updated_at) select id, enable, enable_ssl, host, port, connect_dn, connect_secret_password, base_dn, user_name_rdn_key, user_email_rdn_key, created_at, updated_at from $database.global_configuration_ldap where deleted_at is null;

  # import oauth2_configurations
  insert ignore into oauth2_configurations (uid, enable_oauth2, client_id, client_secret, client_host, server_auth_url, server_token_url, server_user_id_url, scopes, access_token_tag, user_id_tag, login_tip, created_at, updated_at) select id, enable_oauth2, client_id, client_secret, client_host, server_auth_url, server_token_url, server_user_id_url, scopes, access_token_tag, user_id_tag, login_tip, created_at, updated_at from $database.global_configuration_oauth2 where deleted_at is null;

  # import basic_configs
  insert ignore into basic_configs (uid, title, logo, created_at, updated_at) select gcp.id, title, logo, gcp.created_at, gcp.updated_at from $database.global_configuration_personalise as gcp, $database.global_configuration_logo;

  # import database_source_services
  insert ignore into database_source_services (uid, name, source, version, url, db_type, project_uid, cron_express, last_sync_success_time, extra_parameters, created_at, updated_at) select concat(sit.id, p.id), concat(sit.source, sit.id, p.id), sit.source, sit.version, sit.url, sit.db_type, if(p.id=1, 700300, p.id), sit.sync_instance_interval, sit.last_sync_success_time, json_object('sqle_config', json_object('rule_template_id', cast(rt.id as char), 'rule_template_name', rt.name, 'sql_query_config', json_object('audit_enabled', false, 'max_pre_query_rows', 0, 'query_timeout_second', 0, 'allow_query_when_less_than_audit_level', \"\"))) as extra_parameters, sit.created_at, sit.updated_at  from $database.sync_instance_tasks sit join $database.rule_templates rt on sit.rule_template_id = rt.id, $database.projects p where sit.deleted_at is null and p.deleted_at is null;

  # import db_services
  insert ignore into db_services (uid, name, db_type, db_host, db_port, db_user, db_password, \`desc\`, business, additional_params, source, project_uid, maintenance_period, extra_parameters, created_at, updated_at) select s.id, s.name, s.db_type, db_host, db_port, db_user, db_password, s.\`desc\`, 'dms' as business, additional_params, source, if(s.project_id=1, 700300, s.project_id), maintenance_period, json_object('sqle_config', json_object('rule_template_id', cast(rt.id as char), 'rule_template_name', rt.name, 'sql_query_config', json_object('audit_enabled', json_extract(s.sql_query_config, '$.audit_enabled'), 'max_pre_query_rows', json_extract(s.sql_query_config, '$.max_pre_query_rows'), 'query_timeout_second', json_extract(s.sql_query_config, '$.query_timeout_second'), 'allow_query_when_less_than_audit_level', json_extract(s.sql_query_config, '$.allow_query_when_less_than_audit_level')))) as extra_parameters, s.created_at, s.updated_at from $database.instances s join $database.instance_rule_template irt on s.id = irt.instance_id join $database.rule_templates rt on irt.rule_template_id = rt.id where s.deleted_at is null;

  # import roles
  insert ignore roles (uid, name, description, stat, created_at, updated_at) select id, name, \`desc\`, stat, created_at, updated_at from $database.roles where deleted_at is null;

  #import role_op_permissions
  insert ignore into role_op_permissions (role_uid, op_permission_uid) select role_id, case when op_code=20100 then '700007' when op_code=20200 then '700003' when op_code=20300 then '700004' when op_code=20400 then '700006' when op_code=30100 then '700008' when op_code=30200 then '700009' when op_code=40100 then '700010' else '' end as op_code from $database.role_operations where deleted_at is null;

  # import user_op_permissions
  insert ignore into user_op_permissions (user_uid, op_permission_uid) select if(user_id=1, '700200', user_id), case when permission_code=1 then '700001' else '' end as permission_code from $database.management_permissions where deleted_at is null;

  # clear
  delete from member_role_op_ranges;
  delete from members;

  # import members
  insert ignore into members(uid, user_uid, project_uid, created_at, updated_at) select @rownum:=@rownum+1, if(user_id=1, '700200', user_id), project_id, now(), now() from (select @rownum:=0) a, (select if(project_id=1, 700300, project_id) project_id, user_id from $database.project_user union select if(project_id=1, 700300, project_id) project_id, user_id from $database.project_manager) union_member;

  # import member_role_op_ranges
  insert ignore into member_role_op_ranges (member_uid, role_uid, op_range_type, range_uids) select m.uid, pmr.role_id, 'db_service' as op_range_type, pmr.instance_id from $database.project_member_roles pmr join $database.roles r on pmr.role_id = r.id and r.deleted_at is null join $database.instances i on pmr.instance_id = i.id join members m on m.user_uid = if(pmr.user_id=1, '700200', pmr.user_id) and m.project_uid = if(i.project_id=1, '700300', i.project_id) where pmr.deleted_at is null;

  # import member_role_op_ranges
  insert ignore into member_role_op_ranges (member_uid, role_uid, op_range_type, range_uids) select m.uid, '700400' as role_id, 'project' as op_range_type, if(pm.project_id=1, '700300', pm.project_id) from $database.project_manager pm join members m on m.user_uid = if(pm.user_id=1, '700200', pm.user_id) and m.project_uid = if(pm.project_id=1, '700300', pm.project_id);

  # generate members per project
  select @member_rownum:=(select count(*) from members);
  insert ignore into members(uid, user_uid, project_uid, created_at, updated_at) select @rownum:=@rownum+1, 700200 as user_id, project_id, now(), now() from (select @rownum:=@member_rownum) a, (select uid project_id from projects) b;
  insert ignore into member_role_op_ranges(member_uid, role_uid, op_range_type, range_uids) select @rownum:=@rownum+1, '700400' as role_uid, 'project' as op_range_type, project_id from (select @rownum:=@member_rownum) a, (select uid project_id from projects) b;

  # import member_groups
  insert ignore into member_groups (uid, project_uid, name, created_at, updated_at) select @rownum:=@rownum+1, project_id, user_group_id as name, now(), now() from (select @rownum:=0) a, (select if(project_id=1, '700300', project_id) project_id, user_group_id from $database.project_user_group) member_group;

  # import member_group_users
  insert ignore into member_group_users (member_group_uid, user_uid) select mg.uid, if(ugu.user_id=1, '700200', ugu.user_id) from $database.project_user_group pug join $database.user_group_users ugu on pug.user_group_id = ugu.user_group_id join member_groups mg on mg.project_uid = if(pug.project_id=1, '700300', pug.project_id) and mg.name = pug.user_group_id;

  # import member_group_role_op_ranges
  insert ignore into member_group_role_op_ranges (member_group_uid, role_uid, op_range_type, range_uids) select mg.uid, pmgr.role_id, 'db_service' as op_range_type, pmgr.instance_id from $database.project_member_group_roles pmgr join $database.project_user_group pug on pmgr.user_group_id = pug.user_group_id join member_groups mg on mg.project_uid = if(pug.project_id=1, '700300', pug.project_id) and mg.name = pug.user_group_id where pmgr.deleted_at is null;
  "
}

funMigrateSQLE(){
  mysql -h "$d_host" -P "$d_port" -u"$d_user" -p"$d_pwd" -v -e "

  # 切换数据库到sqle
  use sqle;

  # import audit_plan_report_sqls_v2
  insert ignore into audit_plan_report_sqls_v2 select * from $database.audit_plan_report_sqls_v2;

  # import audit_plan_reports_v2
  insert ignore into audit_plan_reports_v2 select * from $database.audit_plan_reports_v2;

  # import audit_plan_sqls_v2
  insert ignore into audit_plan_sqls_v2 select * from $database.audit_plan_sqls_v2;

  # import audit_plans
  insert ignore into audit_plans (id, created_at, updated_at, deleted_at, project_id, name, cron_expression, db_type, token, instance_name, create_user_id, instance_database, type, rule_template_name, params, notify_interval, notify_level, enable_email_notify, enable_web_hook_notify, web_hook_url, web_hook_template) select id, created_at, updated_at, deleted_at, if(project_id=1, '700300', project_id), name, cron_expression, db_type, token, instance_name, if(create_user_id=1, 700200, create_user_id), instance_database, type, rule_template_name, params, notify_interval, notify_level, enable_email_notify, enable_web_hook_notify, web_hook_url, web_hook_template from $database.audit_plans;

  # import black_list_audit_plan_sqls
  insert ignore into black_list_audit_plan_sqls select * from $database.black_list_audit_plan_sqls;

  # import company_notices
  insert ignore into company_notices select * from $database.company_notices;

  # import custom_rules
  insert ignore into custom_rules select * from $database.custom_rules;

  # import ding_talk_instances
  insert ignore into ding_talk_instances select * from $database.ding_talk_instances;

  # import execute_sql_detail
  # fix wrong sql: insert ignore into execute_sql_detail select * from $database.execute_sql_detail;
  insert ignore into execute_sql_detail(id, created_at, updated_at, deleted_at, task_id, number, content, description, start_binlog_file, start_binlog_pos, end_binlog_file, end_binlog_pos, row_affects, exec_status, exec_result, \`schema\`, source_file, audit_status, audit_results, audit_fingerprint, audit_level) select id, created_at, updated_at, deleted_at, task_id, number, content, description, start_binlog_file, start_binlog_pos, end_binlog_file, end_binlog_pos, row_affects, exec_status, exec_result, \`schema\`, '' AS source_file, audit_status, audit_results, audit_fingerprint, audit_level from $database.execute_sql_detail;

  # import feishu_instances
  insert ignore into feishu_instances select * from $database.feishu_instances;

  # import operation_records
  insert ignore into operation_records select * from $database.operation_records;

  # import rollback_sql_detail
  # fix wrong sql: insert ignore into rollback_sql_detail select * from $database.rollback_sql_detail;
  insert ignore into rollback_sql_detail(id, created_at, updated_at, deleted_at, task_id, number, content, description, start_binlog_file, start_binlog_pos, end_binlog_file, end_binlog_pos, row_affects, exec_status, exec_result, \`schema\`, source_file, execute_sql_id) select id, created_at, updated_at, deleted_at, task_id, number, content, description, start_binlog_file, start_binlog_pos, end_binlog_file, end_binlog_pos, row_affects, exec_status, exec_result, \`schema\`, '' AS source_file, execute_sql_id from $database.rollback_sql_detail;
  # import rule_knowledge
  insert ignore into rule_knowledge select * from $database.rule_knowledge;

  # import rule_template_custom_rules
  insert ignore into rule_template_custom_rules select * from $database.rule_template_custom_rules;

  # import rule_template_rule
  truncate table rule_template_rule;
  insert ignore into rule_template_rule select * from $database.rule_template_rule;

  # import rule_templates
  truncate table rule_templates;
  insert ignore into rule_templates(id, created_at, updated_at, deleted_at, project_id, name, \`desc\`, db_type) select id, created_at, updated_at, deleted_at, if(project_id=1, '700300', project_id), name, \`desc\`, db_type from $database.rule_templates;

  # import rules
  truncate table rules;
  insert ignore into rules select * from $database.rules;

  # import sql_audit_records
  insert ignore into sql_audit_records(id, created_at, updated_at, deleted_at, project_id, creator_id, audit_record_id, tags, task_id) select id, created_at, updated_at, deleted_at, if(project_id=1, '700300', project_id), if(creator_id=1, '700200', creator_id), audit_record_id, tags, task_id from $database.sql_audit_records;

  # import sql_manage_sql_audit_records
  insert ignore into sql_manage_sql_audit_records select * from $database.sql_manage_sql_audit_records;

  # import sql_manages
  insert ignore into sql_manages (id, created_at, updated_at, deleted_at, sql_fingerprint, proj_fp_source_inst_schema_md5, sql_text, source, audit_level, audit_results, fp_count, first_appear_timestamp, last_receive_timestamp, instance_name, schema_name, status, remark, project_id, audit_plan_id) select id, created_at, updated_at, deleted_at, sql_fingerprint, proj_fp_source_inst_schema_md5, sql_text, source, audit_level, audit_results, fp_count, first_appear_timestamp, last_receive_timestamp, instance_name, schema_name, status, remark, if(project_id=1, '700300', project_id), audit_plan_id from $database.sql_manages;
  update sql_manages sm join (select group_concat(if(u.id=1, '700200', u.id)) assignees, sma.sql_manage_id from $database.sql_manage_assignees sma join $database.users u on sma.user_id = u.id group by sma.sql_manage_id) sma on sma.sql_manage_id = sm.id set sm.assignees = sma.assignees;

  # import sql_query_execution_sqls
  insert ignore into sql_query_execution_sqls select * from $database.sql_query_execution_sqls;

  # import sql_whitelist
  insert ignore into sql_whitelist(id, created_at, updated_at, deleted_at, project_id, value, \`desc\`, message_digest, match_type) select id, created_at, updated_at, deleted_at, if(project_id=1, '700300', project_id), value, \`desc\`, message_digest, match_type from $database.sql_whitelist;

  # import system_variables
  insert ignore into system_variables select * from $database.system_variables;

  # import task_groups
  insert ignore into task_groups select * from $database.task_groups;

  # import tasks
  insert ignore into tasks(id, created_at, updated_at, deleted_at, instance_id, instance_schema, pass_rate, score, audit_level, sql_source, db_type, status, group_id, create_user_id, exec_start_at, exec_end_at) select id, created_at, updated_at, deleted_at, instance_id, instance_schema, pass_rate, score, audit_level, sql_source, db_type, status, group_id, if(create_user_id=1, 700200, create_user_id), exec_start_at, exec_end_at from $database.tasks;

  # import workflow_instance_records
  insert ignore into workflow_instance_records(id, created_at, updated_at, deleted_at, task_id, workflow_record_id, instance_id, scheduled_at, schedule_user_id, is_sql_executed, execution_user_id) select id, created_at, updated_at, deleted_at, task_id, workflow_record_id, instance_id, scheduled_at, if(schedule_user_id=1, '700200', schedule_user_id), is_sql_executed, if(execution_user_id=1, '700200', execution_user_id) from $database.workflow_instance_records;
  update workflow_instance_records wir join (select group_concat(if(u.id=1, '700200', u.id)) assignees, workflow_instance_record_id from $database.workflow_instance_record_user wiru join $database.users u on wiru.user_id = u.id group by wiru.workflow_instance_record_id) wiru on wir.id = wiru.workflow_instance_record_id set wir.execution_assignees = wiru.assignees;

  # import workflow_record_history
  insert ignore into workflow_record_history select * from $database.workflow_record_history;

  # import workflow_records
  insert ignore into workflow_records select * from $database.workflow_records;

  # import workflow_step_templates
  truncate table workflow_step_templates;
  insert ignore into workflow_step_templates(id, created_at, updated_at, deleted_at, step_number, workflow_template_id, type, \`desc\`, approved_by_authorized, execute_by_authorized) select id, created_at, updated_at, deleted_at, step_number, workflow_template_id, type, \`desc\`, approved_by_authorized, execute_by_authorized from $database.workflow_step_templates;
  update workflow_step_templates wst join (select group_concat(if(u.id=1, '700200', u.id)) assignees, workflow_step_template_id from $database.workflow_step_template_user wstu join $database.users u on wstu.user_id = u.id group by wstu.workflow_step_template_id) wstu on wst.id = wstu.workflow_step_template_id set wst.users = wstu.assignees;

  # import workflow_steps
  insert ignore into workflow_steps(id, created_at, updated_at, deleted_at, operation_user_id, operate_at, workflow_id, workflow_record_id, workflow_step_template_id, state, reason) select ws.id, ws.created_at, ws.updated_at, ws.deleted_at, if(ws.operation_user_id=1, '700200', ws.operation_user_id), ws.operate_at, w.workflow_id, ws.workflow_record_id, ws.workflow_step_template_id, ws.state, ws.reason from $database.workflow_steps ws join $database.workflows w on ws.workflow_id = w.id;
  update workflow_steps ws join (select group_concat(if(u.id=1, '700200', u.id)) assignees, workflow_step_id from $database.workflow_step_user wsu join $database.users u on wsu.user_id = u.id group by wsu.workflow_step_id) wsu on ws.id = wsu.workflow_step_id set ws.assignees = wsu.assignees;

  # import workflow_templates
  truncate table workflow_templates;
  insert ignore into workflow_templates(id, created_at, updated_at, deleted_at, project_id, name, \`desc\`, allow_submit_when_less_audit_level) select wt.id, wt.created_at, wt.updated_at, wt.deleted_at, if(p.id = 1, '700300', p.id) project_id, wt.name, wt.\`desc\`, wt.allow_submit_when_less_audit_level from $database.workflow_templates wt join $database.projects p on p.workflow_template_id = wt.id;

  # import workflows
  insert ignore into workflows (id, created_at, updated_at, deleted_at, subject, workflow_id, \`desc\`, create_user_id, workflow_record_id, project_id) select id, created_at, updated_at, deleted_at, subject, workflow_id, \`desc\`, if(create_user_id=1, '700200', create_user_id), workflow_record_id, if(project_id=1, '700300', project_id) from $database.workflows;
  "
}

if [[ "$edition" == 'enterprise' ]]; then
    mysql -h "$d_host" -P "$d_port" -u"$d_user" -p"$d_pwd" -e "

    # 切换数据库到sqle
    use sqle;

    # import cluster_leader
    insert ignore into cluster_leader select * from $database.cluster_leader;

    # import cluster_node_info
    insert ignore into cluster_node_info select * from $database.cluster_node_info;
    "
fi

funMigrateDMS
#
funMigrateSQLE

