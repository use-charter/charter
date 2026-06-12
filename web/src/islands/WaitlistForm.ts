export function validateEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

export type SubmitResult =
  | { success: true; message: string }
  | { success: false; error: string };

export async function submitWaitlist(email: string): Promise<SubmitResult> {
  try {
    const res = await fetch('/api/waitlist', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email }),
    });
    const data = await res.json();
    if (res.ok) {
      return { success: true, message: data.message || 'Check your email!' };
    }
    return { success: false, error: data.error || 'Something went wrong' };
  } catch {
    return { success: false, error: 'Something went wrong' };
  }
}

export function showToast(message: string, type: 'success' | 'error'): void {
  const existing = document.getElementById('waitlist-toast');
  if (existing) existing.remove();

  const toast = document.createElement('div');
  toast.id = 'waitlist-toast';
  toast.setAttribute('role', 'status');
  toast.setAttribute('aria-live', 'polite');
  toast.setAttribute('aria-atomic', 'true');
  toast.className = `toast toast--${type}`;
  toast.textContent = message;
  document.body.appendChild(toast);

  setTimeout(() => toast.remove(), 4000);
}

export function initWaitlistForm(): void {
  const form = document.getElementById('waitlist-form') as HTMLFormElement | null;
  if (!form) return;

  const input = document.getElementById('waitlist-email') as HTMLInputElement | null;
  const errorSpan = document.getElementById('email-error') as HTMLElement | null;
  const submitBtn = document.getElementById('waitlist-submit') as HTMLButtonElement | null;
  if (!input || !errorSpan || !submitBtn) return;

  input.addEventListener('blur', () => {
    if (input.value && !validateEmail(input.value)) {
      errorSpan.textContent = 'Please enter a valid email address.';
      input.setAttribute('aria-invalid', 'true');
    }
  });

  input.addEventListener('input', () => {
    errorSpan.textContent = '';
    input.removeAttribute('aria-invalid');
  });

  form.addEventListener('submit', async (e) => {
    e.preventDefault();

    if (!validateEmail(input.value)) {
      errorSpan.textContent = 'Please enter a valid email address.';
      input.setAttribute('aria-invalid', 'true');
      input.focus();
      return;
    }

    submitBtn.disabled = true;
    errorSpan.textContent = '';

    const result = await submitWaitlist(input.value);

    submitBtn.disabled = false;

    if (result.success) {
      showToast(result.message, 'success');
      form.reset();
      // Flash the success checkmark on the submit button (mirrors CopyButton).
      submitBtn.classList.add('is-done');
      setTimeout(() => submitBtn.classList.remove('is-done'), 2000);
    } else {
      showToast(result.error, 'error');
    }
  });
}
