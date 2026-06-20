export function validateEmail(email: string): boolean {
	return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

export type SubmitResult =
	| { success: true; message: string }
	| { success: false; error: string };

export async function submitWaitlist(
	email: string,
	company = "",
): Promise<SubmitResult> {
	try {
		const res = await fetch("/api/waitlist", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			// `company` is the honeypot — empty for real users, carried so the server can drop bots.
			body: JSON.stringify({ email, company }),
		});
		const data = await res.json();
		if (res.ok) {
			return { success: true, message: data.message || "Check your email!" };
		}
		return { success: false, error: data.error || "Something went wrong" };
	} catch {
		return { success: false, error: "Something went wrong" };
	}
}

export function initWaitlistForm(): void {
	const form = document.getElementById(
		"waitlist-form",
	) as HTMLFormElement | null;
	if (!form) return;

	const input = document.getElementById(
		"waitlist-email",
	) as HTMLInputElement | null;
	const honeypot = document.getElementById(
		"waitlist-company",
	) as HTMLInputElement | null;
	const errorSpan = document.getElementById(
		"email-error",
	) as HTMLElement | null;
	const submitBtn = document.getElementById(
		"waitlist-submit",
	) as HTMLButtonElement | null;
	const successText = form.querySelector(
		".ck-foot__success-text",
	) as HTMLElement | null;
	if (!input || !errorSpan || !submitBtn || !successText) return;

	input.addEventListener("blur", () => {
		if (input.value && !validateEmail(input.value)) {
			errorSpan.textContent = "Please enter a valid email address.";
			input.setAttribute("aria-invalid", "true");
		}
	});

	input.addEventListener("input", () => {
		errorSpan.textContent = "";
		input.removeAttribute("aria-invalid");
	});

	form.addEventListener("submit", async (e) => {
		e.preventDefault();

		if (!validateEmail(input.value)) {
			errorSpan.textContent = "Please enter a valid email address.";
			input.setAttribute("aria-invalid", "true");
			input.focus();
			return;
		}

		submitBtn.disabled = true;
		errorSpan.textContent = "";

		const result = await submitWaitlist(input.value, honeypot?.value ?? "");

		submitBtn.disabled = false;

		if (result.success) {
			// The form row morphs into an inline confirmation, then reverts.
			successText.textContent = result.message;
			form.reset();
			form.classList.add("is-sent");
			setTimeout(() => form.classList.remove("is-sent"), 4000);
		} else {
			errorSpan.textContent = result.error;
		}
	});
}
