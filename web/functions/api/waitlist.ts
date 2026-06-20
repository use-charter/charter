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

const FROM = "Charter <updates@use-charter.dev>";
const SITE_HOST = "use-charter.dev";
const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const SUCCESS = { success: true, message: "Thanks — you're on the list." };
const escapeHtml = (s: string) =>
	s.replace(
		/[<>&"']/g,
		(c) =>
			({ "<": "&lt;", ">": "&gt;", "&": "&amp;", '"': "&quot;", "'": "&#39;" })[
				c
			] ?? c,
	);
const json = (body: unknown, status = 200) =>
	new Response(JSON.stringify(body), {
		status,
		headers: { "Content-Type": "application/json" },
	});

// Same-origin guard: the form posts from the site itself, so a request whose
// Origin/Referer is missing or foreign is not the form (curl floods, cross-site
// abuse) and is rejected before any Resend call.
const sameOrigin = (request: Request): boolean => {
	const ref =
		request.headers.get("Origin") ?? request.headers.get("Referer") ?? "";
	try {
		return new URL(ref).host === SITE_HOST;
	} catch {
		return false;
	}
};

export const onRequestPost = async (context: {
	request: Request;
	env: Env;
}): Promise<Response> => {
	const { request, env } = context;

	if (!sameOrigin(request)) {
		return json({ success: false, error: "Invalid request." }, 403);
	}

	let email = "";
	let honeypot = "";
	try {
		const data = (await request.json()) as {
			email?: unknown;
			company?: unknown;
		};
		email = String(data?.email ?? "").trim();
		honeypot = String(data?.company ?? "").trim();
	} catch {
		return json({ success: false, error: "Invalid request." }, 400);
	}

	// Honeypot: `company` is a hidden field no human fills. A bot that completes it
	// gets a success-shaped response but no email is sent, so it can't tell it was
	// filtered (and won't adapt).
	if (honeypot) {
		return json(SUCCESS);
	}

	if (!EMAIL_RE.test(email) || email.length > 254) {
		return json(
			{ success: false, error: "Please enter a valid email address." },
			400,
		);
	}

	try {
		const res = await fetch("https://api.resend.com/emails", {
			method: "POST",
			headers: {
				Authorization: `Bearer ${env.RESEND_API_KEY}`,
				"Content-Type": "application/json",
			},
			body: JSON.stringify({
				from: FROM,
				to: [env.WAITLIST_TO],
				reply_to: email,
				subject: "New Charter updates subscriber",
				html: `<p><strong>${escapeHtml(email)}</strong> subscribed for Charter product updates.</p>`,
			}),
		});
		if (!res.ok) {
			return json(
				{ success: false, error: "Something went wrong. Please try again." },
				502,
			);
		}
	} catch {
		return json(
			{ success: false, error: "Something went wrong. Please try again." },
			500,
		);
	}

	return json(SUCCESS);
};
