import { NextRequest, NextResponse } from 'next/server'
import { Resend } from 'resend'

const resend = new Resend(process.env.RESEND_API_KEY)

export async function POST(req: NextRequest) {
  const { email } = await req.json()

  if (!email || typeof email !== 'string' || !email.includes('@')) {
    return NextResponse.json({ error: 'Invalid email' }, { status: 400 })
  }

  await resend.emails.send({
    from: 'Daco <daco@dacolabs.com>',
    to: 'daco@dacolabs.com',
    subject: `New early access request: ${email}`,
    text: `${email} requested early access from the landing page.`,
  })

  return NextResponse.json({ ok: true })
}
