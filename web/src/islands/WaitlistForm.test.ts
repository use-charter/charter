import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { initWaitlistForm, showToast, submitWaitlist, validateEmail } from './WaitlistForm';

describe('validateEmail', () => {
  it('rejects plain string with no @ or domain', () => {
    expect(validateEmail('notanemail')).toBe(false);
  });

  it('accepts valid email', () => {
    expect(validateEmail('user@example.com')).toBe(true);
  });

  it('rejects email with space before @', () => {
    expect(validateEmail('user @example.com')).toBe(false);
  });

  it('rejects email missing domain', () => {
    expect(validateEmail('user@')).toBe(false);
  });

  it('rejects empty string', () => {
    expect(validateEmail('')).toBe(false);
  });

  it('accepts subdomain email', () => {
    expect(validateEmail('user@mail.example.com')).toBe(true);
  });
});

describe('submitWaitlist', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('sends POST to /api/waitlist with email', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ success: true, message: 'Check your email!' }),
    });
    vi.stubGlobal('fetch', fetchMock);

    await submitWaitlist('user@example.com');

    expect(fetchMock).toHaveBeenCalledWith('/api/waitlist', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: 'user@example.com' }),
    });
  });

  it('returns success result on HTTP 200', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ success: true, message: 'Check your email!' }),
    }));

    const result = await submitWaitlist('user@example.com');
    expect(result.success).toBe(true);
    if (result.success) expect(result.message).toBe('Check your email!');
  });

  it('returns error result on HTTP 400', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: false,
      json: async () => ({ error: 'Already subscribed' }),
    }));

    const result = await submitWaitlist('user@example.com');
    expect(result.success).toBe(false);
    if (!result.success) expect(result.error).toBe('Already subscribed');
  });

  it('returns generic error on network failure', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Network error')));

    const result = await submitWaitlist('user@example.com');
    expect(result.success).toBe(false);
    if (!result.success) expect(result.error).toBe('Something went wrong');
  });

  it('uses fallback message when server returns success without a message', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({}),
    }));

    const result = await submitWaitlist('user@example.com');
    expect(result.success).toBe(true);
    if (result.success) expect(result.message).toBe('Check your email!');
  });

  it('uses fallback error message when server returns no error field', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: false,
      json: async () => ({}),
    }));

    const result = await submitWaitlist('user@example.com');
    expect(result.success).toBe(false);
    if (!result.success) expect(result.error).toBe('Something went wrong');
  });
});

describe('initWaitlistForm DOM', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    document.body.innerHTML = `
      <form id="waitlist-form" novalidate>
        <label for="waitlist-email">Email address</label>
        <input type="email" id="waitlist-email" name="email" />
        <span id="email-error" role="alert" aria-atomic="true"></span>
        <button type="submit" id="waitlist-submit">Notify me</button>
      </form>
      <div id="toast-container"></div>
    `;
    initWaitlistForm();
  });

  it('shows error on blur with invalid email', () => {
    const input = document.getElementById('waitlist-email') as HTMLInputElement;
    const errorSpan = document.getElementById('email-error') as HTMLElement;

    input.value = 'bademail';
    input.dispatchEvent(new Event('blur'));

    expect(errorSpan.textContent).not.toBe('');
  });

  it('clears error on input after fixing email', () => {
    const input = document.getElementById('waitlist-email') as HTMLInputElement;
    const errorSpan = document.getElementById('email-error') as HTMLElement;

    input.value = 'bademail';
    input.dispatchEvent(new Event('blur'));
    expect(errorSpan.textContent).not.toBe('');

    input.value = 'good@example.com';
    input.dispatchEvent(new Event('input'));
    expect(errorSpan.textContent).toBe('');
  });

  it('disables submit button during POST', async () => {
    let resolvePost!: (v: unknown) => void;
    vi.stubGlobal('fetch', vi.fn().mockImplementation(() =>
      new Promise(res => { resolvePost = res; })
    ));

    const input = document.getElementById('waitlist-email') as HTMLInputElement;
    const submitBtn = document.getElementById('waitlist-submit') as HTMLButtonElement;
    const form = document.getElementById('waitlist-form') as HTMLFormElement;

    input.value = 'user@example.com';
    form.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));

    await new Promise(r => setTimeout(r, 0));
    expect(submitBtn.disabled).toBe(true);

    resolvePost({ ok: true, json: async () => ({ success: true, message: 'Check your email!' }) });
  });

  it('does not flag an empty value on blur', () => {
    const input = document.getElementById('waitlist-email') as HTMLInputElement;
    const errorSpan = document.getElementById('email-error') as HTMLElement;

    input.value = '';
    input.dispatchEvent(new Event('blur'));

    expect(errorSpan.textContent).toBe('');
  });

  it('blocks submit and shows an error for an invalid email', async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);

    const input = document.getElementById('waitlist-email') as HTMLInputElement;
    const errorSpan = document.getElementById('email-error') as HTMLElement;
    const form = document.getElementById('waitlist-form') as HTMLFormElement;

    input.value = 'bademail';
    form.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
    await Promise.resolve();

    expect(errorSpan.textContent).not.toBe('');
    expect(input.getAttribute('aria-invalid')).toBe('true');
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it('shows a success toast and resets the form on success', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ success: true, message: 'Check your email!' }),
    }));

    const input = document.getElementById('waitlist-email') as HTMLInputElement;
    const submitBtn = document.getElementById('waitlist-submit') as HTMLButtonElement;
    const form = document.getElementById('waitlist-form') as HTMLFormElement;

    input.value = 'user@example.com';
    form.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));

    await vi.waitFor(() => {
      const toast = document.getElementById('waitlist-toast');
      expect(toast?.className).toContain('toast--success');
    });
    expect(input.value).toBe('');
    expect(submitBtn.disabled).toBe(false);
  });

  it('flashes the done state on the submit button on success', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ success: true, message: 'Check your email!' }),
    }));

    const input = document.getElementById('waitlist-email') as HTMLInputElement;
    const submitBtn = document.getElementById('waitlist-submit') as HTMLButtonElement;
    const form = document.getElementById('waitlist-form') as HTMLFormElement;

    input.value = 'user@example.com';
    form.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));

    await vi.waitFor(() => {
      expect(submitBtn.classList.contains('is-done')).toBe(true);
    });
  });

  it('reverts the submit button done state after the timeout', async () => {
    vi.useFakeTimers();
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ success: true, message: 'Check your email!' }),
    }));

    const input = document.getElementById('waitlist-email') as HTMLInputElement;
    const submitBtn = document.getElementById('waitlist-submit') as HTMLButtonElement;
    const form = document.getElementById('waitlist-form') as HTMLFormElement;

    input.value = 'user@example.com';
    form.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));

    await vi.runAllTimersAsync();
    expect(submitBtn.classList.contains('is-done')).toBe(false);
    vi.useRealTimers();
  });

  it('shows an error toast on a failed submit', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: false,
      json: async () => ({ error: 'Already subscribed' }),
    }));

    const input = document.getElementById('waitlist-email') as HTMLInputElement;
    const form = document.getElementById('waitlist-form') as HTMLFormElement;

    input.value = 'user@example.com';
    form.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));

    await vi.waitFor(() => {
      const toast = document.getElementById('waitlist-toast');
      expect(toast?.className).toContain('toast--error');
    });
  });
});

describe('initWaitlistForm guards', () => {
  it('no-ops when the form is absent', () => {
    document.body.innerHTML = '';
    expect(() => initWaitlistForm()).not.toThrow();
  });

  it('no-ops when required fields are missing', () => {
    document.body.innerHTML = '<form id="waitlist-form"></form>';
    expect(() => initWaitlistForm()).not.toThrow();
  });
});

describe('showToast', () => {
  beforeEach(() => {
    document.body.innerHTML = '';
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it('appends a toast with status role, type class, and message', () => {
    showToast('Saved', 'success');

    const toast = document.getElementById('waitlist-toast') as HTMLElement;
    expect(toast).not.toBeNull();
    expect(toast.getAttribute('role')).toBe('status');
    expect(toast.className).toBe('toast toast--success');
    expect(toast.textContent).toBe('Saved');
  });

  it('replaces an existing toast instead of stacking', () => {
    showToast('First', 'success');
    showToast('Second', 'error');

    const toasts = document.querySelectorAll('#waitlist-toast');
    expect(toasts.length).toBe(1);
    expect(toasts[0].textContent).toBe('Second');
    expect((toasts[0] as HTMLElement).className).toBe('toast toast--error');
  });

  it('auto-removes the toast after the timeout', () => {
    vi.useFakeTimers();
    showToast('Bye', 'success');
    expect(document.getElementById('waitlist-toast')).not.toBeNull();

    vi.advanceTimersByTime(4000);
    expect(document.getElementById('waitlist-toast')).toBeNull();
  });
});
