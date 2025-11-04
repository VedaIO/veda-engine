<script lang="ts">
  import { onMount } from 'svelte';
  import { writable } from 'svelte/store';

  interface WebBlocklistItem {
    domain: string;
    title: string;
    iconUrl: string;
  }

  let webBlocklistItems = writable<WebBlocklistItem[]>([]);
  let unblockWebStatus = writable('');

  // This array will hold the domains of the currently selected checkboxes.
  // Svelte's `bind:group` directive will automatically keep this array in sync with the UI.
  let selectedWebsites: string[] = [];

  // Fetches the web blocklist from the backend.
  // It uses the `cache: 'no-cache'` option to prevent the browser from returning a stale list.
  // This is crucial for ensuring that the UI reflects the latest state after an item is removed.
  async function loadWebBlocklist(): Promise<void> {
    const res = await fetch('/api/web-blocklist', { cache: 'no-cache' });
    const data = await res.json();
    if (data && data.length > 0) {
      webBlocklistItems.set(data);
    } else {
      webBlocklistItems.set([]);
    }
  }

  // Removes a single domain from the blocklist.
  async function removeWebBlocklist(domain: string): Promise<void> {
    if (confirm(`Bạn có chắc chắn muốn bỏ chặn ${domain} không?`)) {
      await fetch('/api/web-blocklist/remove', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ domain: domain }),
      });
      loadWebBlocklist();
    }
  }

  // Unblocks all websites that are currently selected in the UI.
  async function unblockSelectedWebsites(): Promise<void> {
    if (selectedWebsites.length === 0) {
      alert('Vui lòng chọn các trang web để bỏ chặn.');
      return;
    }

    // Create an array of fetch promises, one for each selected domain.
    // This allows us to send the removal requests in parallel.
    const removalPromises = selectedWebsites.map(async (domain) => {
      const response = await fetch('/api/web-blocklist/remove', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ domain }),
      });
      // It's important to check if the request was successful. If not, alert the user.
      // This helps in diagnosing backend issues (like the 404 error we encountered).
      if (!response.ok) {
        alert(`Error unblocking ${domain}: ${response.statusText}`);
        throw new Error(`Failed to unblock ${domain}`);
      }
    });

    try {
      // Wait for all the removal requests to complete.
      await Promise.all(removalPromises);
    } catch {
      // If any of the promises fail, the error will be caught here.
      // The individual errors are already alerted, so we just stop execution.
      return;
    }

    unblockWebStatus.set('Đã bỏ chặn: ' + selectedWebsites.join(', '));
    setTimeout(() => {
      unblockWebStatus.set('');
    }, 3000);
    loadWebBlocklist(); // Refresh the list
    selectedWebsites = []; // Clear the selection
  }

  async function clearWebBlocklist(): Promise<void> {
    if (
      confirm('Bạn có chắc chắn muốn xóa toàn bộ danh sách chặn web không?')
    ) {
      await fetch('/api/web-blocklist/clear', { method: 'POST' });
      unblockWebStatus.set('Đã xóa toàn bộ danh sách chặn web.');
      setTimeout(() => {
        unblockWebStatus.set('');
      }, 3000);
      loadWebBlocklist(); // Refresh the list
    }
  }

  async function saveWebBlocklist(): Promise<void> {
    const response = await fetch('/api/web-blocklist/save');
    const blob = await response.blob();
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.style.display = 'none';
    a.href = url;
    a.download = 'procguard_web_blocklist.json';
    document.body.appendChild(a);
    a.click();
    window.URL.revokeObjectURL(url);
  }

  async function loadWebBlocklistFile(event: Event): Promise<void> {
    const file = (event.target as HTMLInputElement).files?.[0];
    if (!file) {
      return;
    }
    const formData = new FormData();
    formData.append('file', file);

    await fetch('/api/web-blocklist/load', {
      method: 'POST',
      body: formData,
    });

    unblockWebStatus.set('Đã tải lên và hợp nhất danh sách chặn web.');
    setTimeout(() => {
      unblockWebStatus.set('');
    }, 3000);
    loadWebBlocklist(); // Refresh the list
  }

  onMount(() => {
    loadWebBlocklist();
  });
</script>

<div class="card mt-3">
  <div class="card-body">
    <h5 class="card-title">Các trang web bị chặn</h5>
    <div class="btn-toolbar" role="toolbar">
      <div class="btn-group me-2" role="group">
        <button
          type="button"
          class="btn btn-primary"
          on:click={unblockSelectedWebsites}
        >
          Bỏ chặn mục đã chọn
        </button>
        <button
          type="button"
          class="btn btn-danger"
          on:click={clearWebBlocklist}
        >
          Xóa toàn bộ
        </button>
      </div>
      <div class="btn-group" role="group">
        <button
          type="button"
          class="btn btn-outline-secondary"
          on:click={saveWebBlocklist}
        >
          Lưu danh sách
        </button>
        <button
          type="button"
          class="btn btn-outline-secondary"
          on:click={() => document.getElementById('load-web-input')?.click()}
        >
          Tải lên danh sách
        </button>
      </div>
    </div>
    <input
      type="file"
      id="load-web-input"
      style="display: none"
      on:change={loadWebBlocklistFile}
    />
    {#if $unblockWebStatus}
      <span id="unblock-web-status" class="form-text">{$unblockWebStatus}</span>
    {/if}
    <div id="web-blocklist-items" class="list-group mt-3">
      {#if $webBlocklistItems.length > 0}
        {#each $webBlocklistItems as item (item.domain)}
          <div
            class="list-group-item d-flex justify-content-between align-items-center"
          >
            <label class="flex-grow-1 mb-0 d-flex align-items-center">
              <input
                class="form-check-input me-2"
                type="checkbox"
                name="blocked-website"
                value={item.domain}
                bind:group={selectedWebsites}
              />
              {#if item.iconUrl}
                <img
                  src={item.iconUrl}
                  class="me-2"
                  style="width: 24px; height: 24px;"
                  alt="Website Icon"
                />
              {:else}
                <div class="me-2" style="width: 24px; height: 24px;"></div>
              {/if}
              <span class="fw-bold me-2">{item.title || item.domain}</span>
            </label>
            <button
              class="btn btn-sm btn-outline-danger"
              on:click={() => removeWebBlocklist(item.domain)}>&times;</button
            >
          </div>
        {/each}
      {:else}
        <div class="list-group-item">Hiện không có trang web nào bị chặn.</div>
      {/if}
    </div>
  </div>
</div>
