#!/usr/bin/env python3
import re, json
from pathlib import Path
from urllib.request import urlopen, Request
from urllib.error import URLError, HTTPError

env = {}
for line in Path('.env').read_text().splitlines():
    line = line.strip()
    if line and not line.startswith('#') and '=' in line:
        k, v = line.split('=', 1)
        env[k.strip()] = v.strip()

api_key = env.get('firebaseApiKey')
email = env.get('email')
password = env.get('password')

if api_key and email and password:
    try:
        url = f'https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key={api_key}'
        body = json.dumps({'email': email, 'password': password, 'returnSecureToken': True}).encode()
        req = Request(url, data=body, headers={'Content-Type': 'application/json'})
        with urlopen(req) as resp:
            env['idToken'] = json.loads(resp.read()).get('idToken', '')
        print(f'[bruno] Firebase token fetched for {email}')
    except HTTPError as e:
        print(f'[bruno] Warning: Firebase sign-in failed ({e.code}): {e.read().decode()}')
    except URLError as e:
        print(f'[bruno] Warning: could not reach Firebase: {e.reason}')
else:
    print('[bruno] Warning: firebaseApiKey / email / password missing in .env — skipping token fetch')

bru_path = Path('api/bruno/Rally Backend API/environments/Rally Dev.bru')
content = bru_path.read_text()

def inject(m):
    key = m.group(1)
    return f'  {key}: {env[key]}' if key in env and env[key] else m.group(0)

bru_path.write_text(re.sub(r'^  (\w+):$', inject, content, flags=re.MULTILINE))
