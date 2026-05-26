import { NextRequest, NextResponse } from 'next/server'

export async function POST(req: NextRequest) {
  const { email } = await req.json()

  if (!email || typeof email !== 'string' || !email.includes('@')) {
    return NextResponse.json({ error: 'Invalid email' }, { status: 400 })
  }

  // TODO: forward to your email service, e.g.:
  //   Resend:     await resend.emails.send({ to: 'daco@dacolabs.com', subject: 'New waitlist signup', text: email })
  //   Loops:      await fetch('https://app.loops.so/api/v1/contacts/create', { ... })
  //   Mailchimp:  POST to your list endpoint
  console.log('[waitlist]', email)

  return NextResponse.json({ ok: true })
}
