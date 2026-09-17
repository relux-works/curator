import json
import sqlite3
import threading
from types import SimpleNamespace
import pytest
import csk_registry.store as module
from test_registry import _health_client, _body

@pytest.mark.parametrize('fail', [False, True])
def test_actual_stalled_refresh(tmp_path, monkeypatch, fail):
    client, store, key, _, _ = _health_client(tmp_path, health_verify_interval=10)
    record=key.sign_record(_body())
    store.append(record,created_at='2026-07-13T02:00:00Z')
    entered,release=threading.Event(),threading.Event()
    original=store._integrity_errors_on
    def stalled(conn):
        entered.set()
        assert release.wait(5)
        if fail:
            raise OSError('injected verifier read failure')
        return original(conn)
    monkeypatch.setattr(store,'_integrity_errors_on',stalled)
    now=[store._health_last_refresh_monotonic]
    monkeypatch.setattr(module,'time',SimpleNamespace(monotonic=lambda:now[0]))
    result=[]
    thread=threading.Thread(target=lambda:result.append(store.refresh_health_verdict()))
    thread.start()
    try:
        assert entered.wait(5)
        now[0]+=20
        assert client.get('/health').status_code==200
        now[0]+=.001
        assert client.get('/health').status_code==503
        with pytest.raises(module.StoreIntegrityError):
            store.append(record,created_at='2026-07-13T02:00:01Z')
    finally:
        release.set()
        thread.join(5)
    assert not thread.is_alive()
    assert result[0].ready is (not fail)
    assert client.get('/health').status_code==(503 if fail else 200)
    monkeypatch.setattr(store,'_integrity_errors_on',original)
    assert store.refresh_health_verdict().ready is (not fail)


def test_canonical_record_corruption(tmp_path):
    client,store,key,_,_=_health_client(tmp_path)
    record=key.sign_record(_body())
    store.append(record,created_at='2026-07-13T02:00:00Z')
    with sqlite3.connect(store.path) as conn:
        bad=dict(record,name='forged')
        conn.execute('UPDATE log SET record_json=? WHERE seq=1',(json.dumps(bad),))
    assert client.get('/health').status_code==200
    assert not store.refresh_health_verdict().ready
    assert client.get('/health').status_code==503
    with pytest.raises(module.StoreIntegrityError):
        store.append(record,created_at='2026-07-13T02:00:01Z')


def test_idempotent_replay_does_not_rewind_verdict(tmp_path):
    _,store,key,_,_=_health_client(tmp_path)
    record=key.sign_record(_body())
    args=dict(auditor_id='a',key='k',body_sha256='ab'*32,created_at='2026-07-13T02:00:00Z',now=1800000000,ttl_seconds=86400)
    first,replay=store.append_idempotent(record,**args)
    assert not replay
    store.append(record,created_at='2026-07-13T02:00:01Z')
    before=store.health_verdict()
    again,replay=store.append_idempotent(record,**args)
    assert replay and first==again
    assert store.health_verdict()==before
    assert (before.verified_log_size,before.verified_head)==store.head()
