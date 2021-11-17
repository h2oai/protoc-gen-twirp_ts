package main

var twirpFileName = "twirp.ts"

// based on https://github.com/larrymyers/protoc-gen-twirp_typescript/blob/master/example/ts_client/twirp.ts
var twirpSource = `
/* tslint:disable */

// This file has been generated by https://github.com/h2oai/protoc-gen-twirp_ts.
// Do not edit.

export interface TwirpErrorJSON {
  code: string
  msg: string
  meta: {
    [index: string]: string
  }
}

export class TwirpError extends Error {
  code: string
  meta: {
    [index: string]: string
  }

  constructor(te: TwirpErrorJSON) {
    super(te.msg)

    this.code = te.code
    this.meta = te.meta
  }
}

export const throwTwirpError = (resp: Response) => {
  return resp.json().then((err: TwirpErrorJSON) => { throw new TwirpError(err) })
}

export const createTwirpRequest = (body: object = {}, headers: object = {}): object => {
  return {
    method: 'POST',
    headers: { ...headers, 'Content-Type': 'application/json' },
    body: JSON.stringify(body || {})
  }
}

export type Fetch = (input: RequestInfo, init?: RequestInit) => Promise<Response>
`
