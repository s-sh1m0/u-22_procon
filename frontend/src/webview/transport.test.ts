import { describe, it, expect } from 'vitest'
import { createHostTransport } from './transport'
import type { HostToWebviewMessage, WebviewToHostMessage } from './protocol'

// 拡張ホスト役: postMessage された内容を記録し、任意の応答を webview 側に返せるようにする。
function setup() {
  const target = new EventTarget()
  const sent: WebviewToHostMessage[] = []
  const transport = createHostTransport({ postMessage: (m) => sent.push(m) }, target)
  const reply = (msg: HostToWebviewMessage) =>
    target.dispatchEvent(new MessageEvent('message', { data: msg }))
  return { transport, sent, reply }
}

describe('createHostTransport', () => {
  it('リクエストを api メッセージとして送り、同じ id の応答で Response を返す', async () => {
    const { transport, sent, reply } = setup()
    const p = transport('/api/jobs/j1')

    expect(sent).toEqual([{ type: 'api', id: 1, method: 'GET', path: '/api/jobs/j1' }])
    reply({ type: 'apiResponse', id: 1, status: 200, body: '{"job_id":"j1"}' })

    const res = await p
    expect(res.status).toBe(200)
    expect(await res.json()).toEqual({ job_id: 'j1' })
  })

  it('method と文字列ボディを中継する', () => {
    const { transport, sent } = setup()
    void transport('/auth/logout', { method: 'POST', body: '{"a":1}' })
    expect(sent[0]).toMatchObject({ method: 'POST', body: '{"a":1}' })
  })

  it('並行リクエストは id で対応づける', async () => {
    const { transport, reply } = setup()
    const a = transport('/api/a')
    const b = transport('/api/b')
    reply({ type: 'apiResponse', id: 2, status: 404, body: '{"message":"nf"}' })
    reply({ type: 'apiResponse', id: 1, status: 200, body: '{}' })
    expect((await a).status).toBe(200)
    expect((await b).status).toBe(404)
  })

  it('ボディを持てないステータスでも Response を作れる', async () => {
    const { transport, reply } = setup()
    const p = transport('/auth/logout', { method: 'POST' })
    reply({ type: 'apiResponse', id: 1, status: 204, body: '' })
    expect((await p).status).toBe(204)
  })

  it('apiError は reject する', async () => {
    const { transport, reply } = setup()
    const p = transport('/api/jobs/j1')
    reply({ type: 'apiError', id: 1, message: 'backend に接続できません' })
    await expect(p).rejects.toThrow('backend に接続できません')
  })

  it('無関係なメッセージは無視する', async () => {
    const { transport, reply } = setup()
    const p = transport('/api/jobs/j1')
    reply({ type: 'apiResponse', id: 99, status: 500, body: '' })
    reply({ type: 'apiResponse', id: 1, status: 200, body: '{}' })
    expect((await p).status).toBe(200)
  })
})
