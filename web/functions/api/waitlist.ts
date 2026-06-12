// Cloudflare Pages Function — receives a product-updates signup and notifies
// WAITLIST_TO by email through the Resend API. The Resend call is server-side,
// so the browser CSP (connect-src 'self') is unaffected; the form posts here
// same-origin.
//
// Cloudflare setup (dashboard → Workers & Pages → your project → Settings):
//   • Variables and Secrets → Add → encrypted secret RESEND_API_KEY = <your Resend key>
//   • Variables and Secrets → Add → variable WAITLIST_TO = <your verified address>
// Resend setup: verify the sending domain (use-charter.dev) so the `from` below
// is allowed.

interface Env {
  RESEND_API_KEY: string;
  WAITLIST_TO: string;
}

const FROM = 'Charter <updates@use-charter.dev>';
const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const escapeHtml = (s: string) =>
  s.replace(/[<>&"']/g, (c) => ({ '<': '&lt;', '>': '&gt;', '&': '&amp;', '"': '&quot;', "'": '&#39;' })[c] ?? c);
const json = (body: unknown, status = 200) =>
  new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });

export const onRequestPost = async (context: { request: Request; env: Env }): Promise<Response> => {
  const { request, env } = context;

  let email = '';
  try {
    const data = (await request.json()) as { email?: unknown };
    email = String(data?.email ?? '').trim();
  } catch {
    return json({ success: false, error: 'Invalid request.' }, 400);
  }

  if (!EMAIL_RE.test(email) || email.length > 254) {
    return json({ success: false, error: 'Please enter a valid email address.' }, 400);
  }

  try {
    const res = await fetch('https://api.resend.com/emails', {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${env.RESEND_API_KEY}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        from: FROM,
        to: [env.WAITLIST_TO],
        reply_to: email,
        subject: 'New Charter updates subscriber',
        html: `<p><strong>${escapeHtml(email)}</strong> subscribed for Charter product updates.</p>`,
      }),
    });
    if (!res.ok) {
      return json({ success: false, error: 'Something went wrong. Please try again.' }, 502);
    }
  } catch {
    return json({ success: false, error: 'Something went wrong. Please try again.' }, 500);
  }

  return json({ success: true, message: "Thanks — you're on the list." });
};
