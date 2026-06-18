#!/usr/bin/env python3
"""E2E for issue #886: ODC password sync after DMS datasource password edit.

Flow:
1. Baseline: Mongo + DMS password aligned, ODC session works.
2. Rotate real Mongo password only -> ODC must fail.
3. Update DMS password to match Mongo -> trigger ODC sync -> ODC works again.
4. Assert connect_connection ciphertext changed while salt unchanged.

Env:
  DMS_URL (default http://10.186.63.138:21004)
  DMS_USER / DMS_PASSWORD (default admin/admin)
  DMS_PROJECT_UID (default 700300)
  DB_SERVICE_NAME (default mongo_trial_27017)
  ODC_CONNECTION_ID (default 1)
  ODC_METADB_* / sshpass for remote metadb checks
  SKIP_MONGO_ROTATE=1 to skip real password rotation
  OUT result json path
"""
from __future__ import annotations

import json
import os
import subprocess
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from http.cookiejar import CookieJar

DMS_URL = os.getenv("DMS_URL", "http://10.186.63.138:21004").rstrip("/")
DMS_USER = os.getenv("DMS_USER", "admin")
DMS_PASSWORD = os.getenv("DMS_PASSWORD", "admin")
PROJECT_UID = os.getenv("DMS_PROJECT_UID", "700300")
DB_SERVICE_NAME = os.getenv("DB_SERVICE_NAME", "mongo_trial_27017")
ORG_ID = int(os.getenv("ODC_ORG_ID", "10001"))
ODC_CONNECTION_ID = int(os.getenv("ODC_CONNECTION_ID", "1"))
ORIG_PWD = os.getenv("MONGO_ORIG_PASSWORD", "pass")
ROTATED_PWD = os.getenv("MONGO_ROTATED_PASSWORD", "pass_e2e_886")
MONGO_HOST = os.getenv("MONGO_HOST", "10.186.60.5")
MONGO_PORT = os.getenv("MONGO_PORT", "27017")
OUT = os.getenv("OUT", "odc_password_sync_e2e_886.json")
SKIP_MONGO_ROTATE = os.getenv("SKIP_MONGO_ROTATE", "") == "1"

METADB_SSH = os.getenv("ODC_METADB_SSH", "root@10.186.63.138")
METADB_SSH_PASS = os.getenv("ODC_METADB_SSH_PASS", " ")
METADB_HOST = os.getenv("ODC_METADB_HOST", "20.20.22.4")
METADB_USER = os.getenv("ODC_METADB_USER", "odc")
METADB_PASS = os.getenv("ODC_METADB_PASS", "OdcMetadb@01")
METADB_NAME = os.getenv("ODC_METADB_NAME", "odc_metadb")


class DmsClient:
    def __init__(self, base: str):
        self.base = base.rstrip("/")
        self.token = ""
        self.jar = CookieJar()
        self.opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(self.jar))

    def api(self, method: str, path: str, body=None, token: str | None = None):
        headers = {"Accept": "application/json", "Content-Type": "application/json"}
        if token or self.token:
            headers["Authorization"] = f"Bearer {token or self.token}"
        data = json.dumps(body).encode() if body is not None else None
        req = urllib.request.Request(self.base + path, data=data, headers=headers, method=method)
        with self.opener.open(req, timeout=120) as resp:
            return json.loads(resp.read().decode())

    def login(self):
        data = self.api("POST", "/v1/dms/sessions", {"session": {"UserName": DMS_USER, "Password": DMS_PASSWORD}})
        if data.get("code") != 0:
            raise RuntimeError(f"DMS login failed: {data}")
        self.token = data["data"]["token"]

    def trigger_odc_sync(self):
        req = urllib.request.Request(f"{self.base}/odc_query/", method="GET", headers={"Cookie": f"dms-token={self.token}"})
        with self.opener.open(req, timeout=120) as resp:
            resp.read()

    def odc_get(self, api_path: str, params=None):
        p = dict(params or {})
        p.setdefault("currentOrganizationId", ORG_ID)
        url = self.base + "/odc_query" + api_path
        if p:
            url += "?" + urllib.parse.urlencode(p)
        req = urllib.request.Request(url, method="GET", headers={"Cookie": f"dms-token={self.token}"})
        with self.opener.open(req, timeout=120) as resp:
            return json.loads(resp.read().decode())

    def odc_post(self, api_path: str, body=None, params=None):
        p = dict(params or {})
        p.setdefault("currentOrganizationId", ORG_ID)
        url = self.base + "/odc_query" + api_path
        if p:
            url += "?" + urllib.parse.urlencode(p)
        data = json.dumps(body or {}).encode()
        req = urllib.request.Request(
            url,
            data=data,
            method="POST",
            headers={"Accept": "application/json", "Content-Type": "application/json", "Cookie": f"dms-token={self.token}"},
        )
        with self.opener.open(req, timeout=120) as resp:
            return json.loads(resp.read().decode())


def mongo_change_password(new_pwd: str, old_pwd: str) -> None:
    js = f"db.changeUserPassword('root', '{new_pwd}')"
    uri = f"mongodb://root:{urllib.parse.quote(old_pwd, safe='')}@{MONGO_HOST}:{MONGO_PORT}/admin"
    cmd = ["docker", "run", "--rm", "mongo:6", "mongosh", uri, "--quiet", "--eval", js]
    out = subprocess.check_output(cmd, text=True, stderr=subprocess.STDOUT)
    if "ok: 1" not in out and "ok:1" not in out.replace(" ", ""):
        raise RuntimeError(f"mongo changeUserPassword failed: {out}")


def metadb_query(sql: str) -> str:
    remote = (
        f"mysql -h{METADB_HOST} -P3306 -u{METADB_USER} -p'{METADB_PASS}' {METADB_NAME} -N -e \"{sql}\""
    )
    cmd = ["sshpass", "-p", METADB_SSH_PASS, "ssh", "-o", "StrictHostKeyChecking=no", METADB_SSH, remote]
    return subprocess.check_output(cmd, text=True).strip()


def read_connection_row() -> dict:
    row = metadb_query(
        "SELECT id, password, salt, cipher, is_password_saved, LENGTH(password) "
        f"FROM connect_connection WHERE id={ODC_CONNECTION_ID}"
    )
    parts = row.split("\t")
    if len(parts) < 6:
        raise RuntimeError(f"unexpected metadb row: {row!r}")
    return {
        "id": int(parts[0]),
        "password": parts[1],
        "salt": parts[2],
        "cipher": parts[3],
        "is_password_saved": parts[4],
        "password_len": int(parts[5]),
    }


def get_db_service(client: DmsClient) -> dict:
    data = client.api(
        "GET",
        f"/v1/dms/projects/{PROJECT_UID}/db_services?filter_by_name={urllib.parse.quote(DB_SERVICE_NAME)}&page_size=1&page_index=1",
    )
    items = data.get("data") or []
    if not items:
        raise RuntimeError(f"db_service {DB_SERVICE_NAME} not found")
    return items[0]


def build_update_payload(svc: dict, password: str) -> dict:
    return {
        "db_service": {
            "name": svc["name"],
            "desc": svc.get("desc") or "",
            "db_type": svc["db_type"],
            "host": svc["host"],
            "port": svc["port"],
            "user": svc["user"],
            "password": password,
            "environment_tag_uid": svc["environment_tag"]["uid"],
            "maintenance_times": svc.get("maintenance_times") or [],
            "additional_params": svc.get("additional_params") or [],
            "sqle_config": svc.get("sqle_config"),
            "enable_backup": svc.get("enable_backup", False),
        }
    }


def update_dms_password(client: DmsClient, svc: dict, password: str) -> None:
    payload = build_update_payload(svc, password)
    ct = client.api("POST", f"/v1/dms/projects/{PROJECT_UID}/db_services/connection", payload)
    if ct.get("code") != 0:
        raise RuntimeError(f"connection test failed: {ct}")
    results = ct.get("data") or []
    if isinstance(results, list):
        for item in results:
            if not item.get("is_connectable"):
                raise RuntimeError(f"connection test not connectable: {item}")
    upd = client.api("PUT", f"/v2/dms/projects/{PROJECT_UID}/db_services/{svc['uid']}", payload)
    if upd.get("code") != 0:
        raise RuntimeError(f"update db_service failed: {upd}")


def try_odc_mongo_session(client: DmsClient) -> tuple[bool, str]:
    try:
        dslist = client.odc_get("/api/v2/datasource/datasources", {"page": 1, "size": 50, "basic": "false"})
    except urllib.error.HTTPError as e:
        return False, f"list datasources HTTP {e.code}"
    except Exception as e:  # noqa: BLE001
        return False, str(e)

    mongo = next(
        (d for d in (dslist.get("data") or {}).get("contents", []) if d.get("type") == "MONGODB"),
        None,
    )
    if not mongo:
        return False, "no MONGODB datasource in ODC"

    ds_id = mongo["id"]
    try:
        dbs = client.odc_get(f"/api/v2/datasource/datasources/{ds_id}/databases", {"page": 1, "size": 20})
        db_items = (dbs.get("data") or {}).get("contents", [])
        if not db_items:
            return False, "no databases"
        db_id = db_items[0]["id"]
        sess = client.odc_post(f"/api/v2/datasource/databases/{db_id}/sessions")
        if not sess.get("successful"):
            return False, f"create session failed: {sess.get('message') or sess}"
        session_id = (sess.get("data") or {}).get("sessionId")
        if not session_id:
            return False, f"empty sessionId: {sess}"
        return True, session_id
    except urllib.error.HTTPError as e:
        body = e.read().decode("utf-8", errors="replace")[:500]
        return False, f"HTTP {e.code}: {body}"
    except Exception as e:  # noqa: BLE001
        return False, str(e)


def restart_dms() -> None:
    stop_cmd = [
        "sshpass", "-p", METADB_SSH_PASS, "ssh", "-o", "StrictHostKeyChecking=no", METADB_SSH,
        "docker exec dms-server-main-ee pkill -f '/opt/dms/bin/dms -conf' || true",
    ]
    start_cmd = [
        "sshpass", "-p", METADB_SSH_PASS, "ssh", "-o", "StrictHostKeyChecking=no", METADB_SSH,
        "docker exec -d dms-server-main-ee /opt/dms/bin/dms -conf /opt/dms/etc/config.yaml",
    ]
    subprocess.call(stop_cmd)
    time.sleep(2)
    subprocess.check_call(start_cmd)
    for _ in range(15):
        try:
            req = urllib.request.Request(
                f"{DMS_URL}/v1/dms/sessions",
                data=json.dumps({"session": {"UserName": DMS_USER, "Password": DMS_PASSWORD}}).encode(),
                headers={"Content-Type": "application/json"},
                method="POST",
            )
            with urllib.request.urlopen(req, timeout=10) as resp:
                if resp.status == 200:
                    return
        except Exception:  # noqa: BLE001
            time.sleep(2)
    raise RuntimeError("DMS did not become ready after restart")


def repair_odc_password_cache(client: DmsClient, svc: dict, password: str) -> None:
    remote = (
        "docker exec mysql-for-dms-main-ee mysql -uroot -pmysqlpass dms -N -e "
        f"\"UPDATE sql_workbench_datasource_caches SET dms_db_service_fingerprint='force-resync' "
        f"WHERE dms_db_service_id='{svc['uid']}';\""
    )
    cmd = ["sshpass", "-p", METADB_SSH_PASS, "ssh", "-o", "StrictHostKeyChecking=no", METADB_SSH, remote]
    subprocess.check_call(cmd)
    update_dms_password(client, svc, password)


def trigger_fresh_odc_sync(client: DmsClient, svc: dict, password: str) -> None:
    repair_odc_password_cache(client, svc, password)
    restart_dms()
    client.login()
    client.trigger_odc_sync()


def ensure_baseline(client: DmsClient, svc: dict) -> None:
    ok, detail = try_odc_mongo_session(client)
    if ok:
        return
    trigger_fresh_odc_sync(client, svc, ORIG_PWD)
    ok, detail = try_odc_mongo_session(client)
    if not ok:
        raise RuntimeError(f"baseline ODC session failed after repair: {detail}")


def main() -> int:
    report: dict = {"steps": [], "ok": False}
    client = DmsClient(DMS_URL)
    client.login()
    svc = get_db_service(client)

    ensure_baseline(client, svc)
    baseline_row = read_connection_row()
    report["baseline_metadb"] = baseline_row
    report["steps"].append({"step": "baseline_odc_session", "ok": True})
    if SKIP_MONGO_ROTATE:
        report["skipped"] = "SKIP_MONGO_ROTATE=1"
        report["ok"] = True
        _write(report)
        print(json.dumps(report, indent=2, ensure_ascii=False))
        return 0

    try:
        mongo_change_password(ROTATED_PWD, ORIG_PWD)
        report["steps"].append({"step": "rotate_mongo_password", "ok": True, "to": ROTATED_PWD})

        ok, detail = try_odc_mongo_session(client)
        report["steps"].append({"step": "odc_after_mongo_rotate_only", "ok": not ok, "detail": detail})
        if ok:
            raise RuntimeError("expected ODC session to fail after mongo password rotated, but succeeded")

        trigger_fresh_odc_sync(client, svc, ROTATED_PWD)
        time.sleep(1)

        after_row = read_connection_row()
        report["after_sync_metadb"] = after_row
        ciphertext_changed = after_row["password"] != baseline_row["password"]
        salt_unchanged = after_row["salt"] == baseline_row["salt"]
        report["steps"].append(
            {
                "step": "metadb_ciphertext",
                "ok": ciphertext_changed and salt_unchanged and after_row["password_len"] == 44,
                "ciphertext_changed": ciphertext_changed,
                "salt_unchanged": salt_unchanged,
                "cipher": after_row["cipher"],
                "is_password_saved": after_row["is_password_saved"],
            }
        )
        if not ciphertext_changed or not salt_unchanged:
            raise RuntimeError(f"metadb assertion failed: baseline={baseline_row} after={after_row}")

        ok, detail = try_odc_mongo_session(client)
        report["steps"].append({"step": "odc_after_dms_password_sync", "ok": ok, "detail": detail})
        if not ok:
            raise RuntimeError(f"ODC session failed after DMS password sync: {detail}")

        report["ok"] = True
        _write(report)
        print(json.dumps(report, indent=2, ensure_ascii=False))
        return 0
    finally:
        try:
            mongo_change_password(ORIG_PWD, ROTATED_PWD)
        except Exception as exc:  # noqa: BLE001
            report.setdefault("cleanup_errors", []).append(f"mongo restore: {exc}")
        try:
            svc = get_db_service(client)
            trigger_fresh_odc_sync(client, svc, ORIG_PWD)
        except Exception as exc:  # noqa: BLE001
            report.setdefault("cleanup_errors", []).append(f"dms restore: {exc}")
        _write(report)


def _write(report: dict) -> None:
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(report, f, indent=2, ensure_ascii=False)


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception as exc:  # noqa: BLE001
        print(f"ERROR: {exc}", file=sys.stderr)
        sys.exit(1)
