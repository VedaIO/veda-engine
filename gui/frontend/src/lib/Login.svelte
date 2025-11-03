<script lang="ts">
  import { onMount } from 'svelte';
  import { writable } from 'svelte/store';

  let title = 'Đăng nhập';
  let hasPassword = false;
  let errorMessage = writable('');
  let password = '';
  let newPassword = '';
  let confirmPassword = '';

  onMount(async () => {
    const response = await fetch('/api/has-password');
    if (response.ok) {
      const data = await response.json();
      hasPassword = data.hasPassword;
    } else {
      errorMessage.set('Lỗi kết nối đến máy chủ.');
    }

    if (hasPassword) {
      title = 'Nhập mật khẩu';
    } else {
      title = 'Tạo mật khẩu';
    }
  });

  async function handleLogin(event: Event) {
    event.preventDefault();

    const response = await fetch('/api/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ password }),
    });
    if (response.ok) {
      const { success } = await response.json();
      if (success) {
        window.location.href = '/';
      } else {
        errorMessage.set('Sai mật khẩu');
      }
    } else {
      errorMessage.set('Lỗi đăng nhập');
    }
  }

  async function handleSetPassword(event: Event) {
    event.preventDefault();

    if (newPassword.trim() === '') {
      errorMessage.set('Mật khẩu không được để trống');
      return;
    }

    if (newPassword !== confirmPassword) {
      errorMessage.set('Mật khẩu không khớp');
      return;
    }

    const response = await fetch('/api/set-password', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ password: newPassword }),
    });

    if (response.ok) {
      window.location.href = '/';
    } else {
      errorMessage.set('Lỗi đặt mật khẩu');
    }
  }
</script>

<main class="login-container">
  <div class="card shadow-sm">
    <div class="card-body p-5">
      <h1 class="card-title text-center mb-4">{title}</h1>
      <form
        on:submit|preventDefault={hasPassword ? handleLogin : handleSetPassword}
        novalidate
      >
        {#if hasPassword}
          <div class="mb-3">
            <label for="password" class="form-label">Mật khẩu</label>
            <input
              type="password"
              class="form-control"
              id="password"
              bind:value={password}
              required
            />
          </div>
        {:else}
          <div class="mb-3">
            <label for="new-password" class="form-label">Mật khẩu mới</label>
            <input
              type="password"
              class="form-control"
              id="new-password"
              bind:value={newPassword}
              required
            />
          </div>
          <div class="mb-3">
            <label for="confirm-password" class="form-label"
              >Xác nhận mật khẩu</label
            >
            <input
              type="password"
              class="form-control"
              id="confirm-password"
              bind:value={confirmPassword}
              required
            />
          </div>
        {/if}
        <button type="submit" class="btn btn-danger w-100 mt-3">
          Tiếp tục
        </button>
      </form>
      {#if $errorMessage}
        <p class="text-danger text-center mt-3">
          {$errorMessage}
        </p>
      {/if}
    </div>
  </div>
</main>

<style>
  .login-container {
    width: 100%;
    max-width: 400px;
  }
</style>
