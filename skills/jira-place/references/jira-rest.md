# Jira REST snippets for playwright-cli

Every snippet runs inside the logged-in Jira page, so `fetch` is same-origin and carries the session cookies. Run each one through a quoted heredoc so the shell never touches the JavaScript:

```bash
JS=$(cat <<'EOF'
…snippet with placeholders filled in…
EOF
)
playwright-cli -s=jira --raw eval "$JS"
```

Placeholders come from `~/.config/jira-placement/config.json`: `__PROJECT__` = `projectKey`, `__EPIC_LINK__` / `__EPIC_NAME__` / `__STORY_POINTS__` = `fields.*`.

Every snippet returns JSON. `loggedIn: false` means the session expired: Jira answered 401, or it redirected to the SSO/login page and returned HTML instead of JSON.

## Session check

```js
async () => {
  const r = await fetch('/rest/api/2/myself', { headers: { Accept: 'application/json' } });
  const t = await r.text();
  try {
    const me = JSON.parse(t);
    return { loggedIn: r.status === 200, status: r.status, me: { name: me.name, displayName: me.displayName } };
  } catch {
    return { loggedIn: false, status: r.status };
  }
}
```

## Resolve fields

```js
async () => {
  const r = await fetch('/rest/api/2/field', { headers: { Accept: 'application/json' } });
  if (r.status === 401) return { loggedIn: false };
  const fields = await r.json();
  const find = (name) => (fields.find((f) => f.name === name) || {}).id || null;
  return { epicLink: find('Epic Link'), epicName: find('Epic Name'), storyPoints: find('Story Points') };
}
```

## Tree

Set `INCLUDE_DONE` to `true` only when the user asks to see Done issues.

```js
async () => {
  const P = '__PROJECT__', EPIC_LINK = '__EPIC_LINK__', POINTS = '__STORY_POINTS__', INCLUDE_DONE = false;
  const H = { Accept: 'application/json', 'Content-Type': 'application/json' };
  const meRes = await fetch('/rest/api/2/myself', { headers: H });
  if (meRes.status !== 200) return { loggedIn: false, status: meRes.status };
  const me = await meRes.json();
  const search = async (jql, fields) => {
    const all = [];
    for (;;) {
      const r = await fetch('/rest/api/2/search', { method: 'POST', headers: H, body: JSON.stringify({ jql, fields, startAt: all.length, maxResults: 100 }) });
      if (!r.ok) throw new Error('search ' + r.status + ': ' + (await r.text()).slice(0, 300));
      const page = await r.json();
      all.push(...page.issues);
      if (page.issues.length === 0 || all.length >= page.total) return all;
    }
  };
  const scope = 'project = ' + JSON.stringify(P) + (INCLUDE_DONE ? '' : ' AND statusCategory != Done');
  const D = ['summary', 'description', 'fixVersions', 'labels', 'assignee', 'status'];
  const [E, T, S] = await Promise.all([
    search(scope + ' AND issuetype = Epic ORDER BY key', D),
    search(scope + ' AND issuetype = Task ORDER BY key', [...D, EPIC_LINK]),
    search(scope + ' AND issuetype = Sub-task ORDER BY key', ['summary', 'assignee', 'status', 'parent', POINTS]),
  ]);
  const who = (f) => (f.assignee ? f.assignee.name : null);
  const status = (f) => f.status.name + (f.status.statusCategory && f.status.statusCategory.key === 'done' ? ' (Done)' : '');
  const cut = (s) => (!s ? '' : s.length > 300 ? s.slice(0, 300) + '…' : s);
  const detail = (i) => ({
    key: i.key,
    summary: i.fields.summary,
    description: cut(i.fields.description),
    fixVersions: (i.fields.fixVersions || []).map((v) => v.name),
    labels: i.fields.labels || [],
    assignee: who(i.fields),
    status: status(i.fields),
  });
  const tasks = new Map(T.map((i) => [i.key, { ...detail(i), own: who(i.fields) === me.name, subtasks: [] }]));
  for (const i of S) {
    const t = i.fields.parent && tasks.get(i.fields.parent.key);
    if (!t) continue; // parent Task is not Open
    t.subtasks.push({ key: i.key, summary: i.fields.summary, assignee: who(i.fields), status: status(i.fields), own: who(i.fields) === me.name, points: i.fields[POINTS] ?? null });
  }
  const epics = new Map(E.map((i) => [i.key, { ...detail(i), tasks: [] }]));
  const tasksWithoutEpic = [];
  for (const i of T) {
    const k = i.fields[EPIC_LINK];
    const e = k && epics.get(k);
    if (e) e.tasks.push(tasks.get(i.key));
    else if (!k) tasksWithoutEpic.push(tasks.get(i.key));
    // A Task whose Epic is Done is dropped.
  }
  return { me: { name: me.name, displayName: me.displayName }, epics: [...epics.values()], tasksWithoutEpic };
}
```

## Issue position

For a Referenced Issue. Replace `__KEY__`.

```js
async () => {
  const EPIC_LINK = '__EPIC_LINK__';
  const H = { Accept: 'application/json' };
  const get = async (key) => {
    const r = await fetch('/rest/api/2/issue/' + encodeURIComponent(key) + '?fields=summary,issuetype,status,assignee,parent,' + EPIC_LINK, { headers: H });
    if (r.status === 401) throw new Error('loggedIn:false');
    if (!r.ok) throw new Error(key + ': ' + r.status + ' ' + (await r.text()).slice(0, 300));
    return r.json();
  };
  const me = await (await fetch('/rest/api/2/myself', { headers: H })).json();
  const i = await get('__KEY__');
  const f = i.fields;
  const type = f.issuetype.name;
  const out = {
    key: i.key, type, inHierarchy: ['Epic', 'Task', 'Sub-task'].includes(type),
    summary: f.summary, status: f.status.name, done: f.status.statusCategory.key === 'done',
    assignee: f.assignee ? f.assignee.name : null, own: !!f.assignee && f.assignee.name === me.name,
    task: null, epic: null,
  };
  let epicKey = f[EPIC_LINK];
  if (type === 'Sub-task' && f.parent) {
    const p = await get(f.parent.key);
    out.task = { key: p.key, summary: p.fields.summary, own: !!p.fields.assignee && p.fields.assignee.name === me.name };
    epicKey = p.fields[EPIC_LINK];
  }
  if (type === 'Epic') out.epic = { key: i.key, summary: f.summary };
  else if (epicKey) out.epic = { key: epicKey, summary: (await get(epicKey)).fields.summary };
  return out;
}
```

## Write

One snippet for every write step. Fill in `__METHOD__`, `__PATH__` and `__BODY__` (a JSON literal) from the table below. `X-Atlassian-Token: no-check` is required: Jira rejects cookie-authenticated writes without it (XSRF check).

```js
async () => {
  const r = await fetch('__PATH__', {
    method: '__METHOD__',
    headers: { Accept: 'application/json', 'Content-Type': 'application/json', 'X-Atlassian-Token': 'no-check' },
    body: JSON.stringify(__BODY__),
  });
  const t = await r.text();
  let body = null;
  try { body = t ? JSON.parse(t) : null; } catch { return { loggedIn: false, status: r.status }; }
  const error = r.ok ? null : [...((body && body.errorMessages) || []), ...Object.entries((body && body.errors) || {}).map(([k, v]) => k + ': ' + v)].join('; ') || t.slice(0, 300);
  return { status: r.status, key: body && body.key, error };
}
```

| op | method | path | body |
|---|---|---|---|
| createEpic | POST | `/rest/api/2/issue` | `{ fields: { project: { key: P }, issuetype: { name: 'Epic' }, summary, [EPIC_NAME]: epicName, assignee: { name: me } } }` |
| createTask | POST | `/rest/api/2/issue` | `{ fields: { project: { key: P }, issuetype: { name: 'Task' }, summary, [EPIC_LINK]: epicKey, assignee: { name: me } } }` |
| createSubtask | POST | `/rest/api/2/issue` | `{ fields: { project: { key: P }, issuetype: { name: 'Sub-task' }, parent: { key: taskKey }, summary, assignee: { name: me } } }` |
| setPoints | PUT | `/rest/api/2/issue/<KEY>` | `{ fields: { [STORY_POINTS]: points } }` |
| linkRelates | POST | `/rest/api/2/issueLink` | `{ type: { name: 'Relates' }, inwardIssue: { key: from }, outwardIssue: { key: to } }` |

In the body, write out the resolved keys and the real field IDs, e.g. `{ fields: { customfield_10008: 'PROJ-1', … } }`. Refs must already be replaced with the keys created by earlier steps. `setPoints` and `linkRelates` return an empty body on success (status 204/201), so no `key` comes back.
