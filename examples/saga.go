package example

func TansIn() {

}

func main() {

}

/*
import { ServiceContext } from "@ivy/api/globals"
import { generateTid } from "@ivy/api/ivy"
import { redis } from "@ivy/api/objects"
import { Saga } from '../../saga-cli'

async function getGlobalTid() {
  return "global-" + await generateTid(redis)
}

export async function transIn(ctx: ServiceContext) {
  let { gid, sid } = ctx.query
  let { amount, transIn } = ctx.data
  console.log(`gid: ${gid} sid: ${sid} transIn ${amount}`)
  if (transIn === 'fail') {
    throw { code: "DATA_ERROR", message: "tranIn error FAIL" }
  }
  return { messag: 'SUCCESS' }
}

export async function transInCompensate(ctx: ServiceContext) {
  let { gid, sid } = ctx.query
  let { amount } = ctx.data
  console.log(`gid: ${gid} sid: ${sid} tranInCompensate ${amount}`)
  return { message: 'SUCCESS' }
}

export async function transOut(ctx: ServiceContext) {
  let { gid, sid } = ctx.query
  let { amount, transIn, transOut } = ctx.data
  console.log(`gid: ${gid} sid: ${sid} transOut ${amount}`)
  if (transOut === 'fail') {
    throw { code: "DATA_ERROR", message: "tranIn error FAIL" }
  }
  return { message: 'SUCCESS' }
}

export async function transOutCompensate(ctx: ServiceContext) {
  let { gid, sid } = ctx.query
  let { amount } = ctx.data
  console.log(`gid: ${gid} sid: ${sid} tranOutCompensate ${amount}`)
  return { message: 'SUCCESS' }
}

export async function transQuery(ctx: ServiceContext) {
  let { gid } = ctx.query
  return { message: 'SUCCESS' }
  // return `gid: ${gid} ` + (Math.random() * 3 > 1 ? 'SUCCESS' : 'FAIL')
}

let host = "http://localhost:4005"

export async function startSagaTrans(ctx: ServiceContext) {
  await ctx.beginTransaction()
  let gid = await getGlobalTid()
  console.log(`order: ${gid} created`)
  let { transIn, transOut } = ctx.data
  let saga = new Saga(`${host}/api/core/saga-svr`, gid)
  saga.add(`${host}/api/core/dtrans/transIn`, `${host}/api/core/dtrans/transInCompensate`, { amount: 30, transIn, transOut })
  saga.add(`${host}/api/core/dtrans/transOut`, `${host}/api/core/dtrans/transOutCompensate`, { amount: 30, transIn, transOut })
  await saga.prepare(`${host}/api/core/dtrans/transQuery`)
  await ctx.trans.commit()
  await saga.commit()
}
*/
