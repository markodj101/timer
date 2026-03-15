<script lang="ts">
  import './app.css';
  import { Events } from "@wailsio/runtime";
  import { GreetService } from "../bindings/changeme";
  import { FileReading } from "../bindings/changeme";
  import { Dialogs } from "@wailsio/runtime";
  import { onMount, onDestroy } from 'svelte';

  
  let path = window.location.pathname;

  
  let soundOptions = [
    { id: 'Ton 1', name: 'Ton 1', url: '/1.mp3' },
    { id: 'Ton 2', name: 'Ton 2', url: '/2.mp3' },
    { id: 'Ton 3', name: 'Ton 3', url: '/3.mp3' },
    { id: 'Ton 4', name: 'Ton 4', url: '/4.mp3' },
    { id: 'Ton 5', name: 'Ton 5', url: '/5.mp3' }
  ];

  let minutes = $state(5);
  let seconds = $state(0);
  let isRunning = $state(false);
  let isPaused = $state(false);
  let volume = $state(80);
  let showMessage = $state(false);
  let displayTime = $state("00:00");
  let displayMessage = $state("");
  let message = $state("Vrijeme je isteklo!");
  let customTone = $state<{ id: string; name: string; url: string } | null>(null);
  let selectedTone = $state(soundOptions[0]);
  let totalSeconds = $derived(minutes * 60 + seconds);
  let leftSeconds = $state(5 * 60);
  let timerInterval: ReturnType<typeof setInterval> | null = null;
  let audioContext: AudioContext;
  let audioBufferMap = new Map<string, AudioBuffer>();
  let currentSource = $state<AudioBufferSourceNode | null>(null);
  let currentGainNode = $state<GainNode | null>(null);
  let params = new URLSearchParams(window.location.search);
  let isDisplay = params.get('view') === 'display';

  $effect(() => {
    if (!isRunning && !isPaused) {
      leftSeconds = totalSeconds;
    }
  });

  function handleStart() {
    if (currentSource) {
      currentSource.stop();
    }
   if(isRunning)return;
   if(leftSeconds <= 0){
    leftSeconds = minutes * 60 + seconds;
   }
    isRunning = true;
    isPaused = false;

    timerInterval = setInterval(() => {
      if (leftSeconds <= 0) {
        clearInterval(timerInterval!);
        timerInterval = null;
        isRunning = false;
        onTimerComplete();        
      } else {
        leftSeconds -= 1;
        Events.Emit("timer-update", formattedTime);
      }
    }, 1000);
  }

  function handlePause() {
    if (timerInterval) {
      clearInterval(timerInterval);
      timerInterval = null;
    }
    isRunning = false;
    isPaused = true;
  }

  function onTimerComplete() {
    playSound(selectedTone);
    Events.Emit("timer-complete", message); 
    isPaused = false; 
  }

  function handleCancel() {
    if (timerInterval) {
      clearInterval(timerInterval);
      timerInterval = null;
    }
    isRunning = false;
    isPaused = false;
    leftSeconds = minutes * 60 + seconds;
    const mins = Math.floor(leftSeconds / 60);
    const secs = leftSeconds % 60;
    const formatted = `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
    Events.Emit("timer-update", formatted);
  }

  onDestroy(() => {
    if (timerInterval) {
      clearInterval(timerInterval);
    }
  });

  onMount(() => {
    audioContext = new AudioContext();
  
    (async () => {
      try {
        await loadSounds();
      } catch (error) {
        console.error('Failed to load sounds:', error);
      }
    })();

    const blockRightClick = (e: MouseEvent) => e.preventDefault();
    window.addEventListener('contextmenu', blockRightClick);

    let unsubTimer: (() => void) | undefined;
    let unsubComplete: (() => void) | undefined;
    let handleKeyDown: ((e: KeyboardEvent) => void) | undefined;

    if (isDisplay) {
      unsubTimer = Events.On("timer-update", (ev) => {
        showMessage = false;
        displayTime = ev.data;
      });
      
      unsubComplete = Events.On("timer-complete", (ev) => {
        showMessage = true;
        displayMessage = ev.data;
      });
    } else {
      handleKeyDown = (e: KeyboardEvent) => {
        const target = e.target as HTMLElement;
        if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA') return;

        if (e.code === 'Space') {
          e.preventDefault(); 
          if (!isRunning && !isPaused) {
            handleStart();
          } else if (isRunning) {
            handlePause();
          } else if (isPaused) {
            handleStart();
          }
        }

        if (e.key === 'r' || e.key === 'R') {
          e.preventDefault();
          handleCancel();
        }
      };

      window.addEventListener('keydown', handleKeyDown);
    }

    return () => {
      window.removeEventListener('contextmenu', blockRightClick);
      
      if (handleKeyDown) {
        window.removeEventListener('keydown', handleKeyDown);
      }
      
      if (unsubTimer) unsubTimer();
      if (unsubComplete) unsubComplete();
      
      if (audioContext && audioContext.state !== 'closed') {
        audioContext.close().catch(err => {
          console.error('Failed to close audio context:', err);
        });
      }
    };
  });

  async function loadSounds() {
    for (let tone of soundOptions) {
      try {
        const response = await fetch(tone.url);
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        const arrayBuffer = await response.arrayBuffer();
        const audioBuffer = await audioContext.decodeAudioData(arrayBuffer);
        audioBufferMap.set(tone.id, audioBuffer);
        console.log(`Loaded ${tone.name}`);
      } catch (err) {
        console.error(`Failed to load ${tone.name}:`, err);
      }
    }
  }

  async function playSound(tone: { id: string; name: string; url: string }) {
    if (currentSource) {
      try { currentSource.stop(); } catch (e) {}
      currentSource = null;
      currentGainNode = null;
    }

    if (!audioContext || audioContext.state === 'closed') {
      audioContext = new AudioContext();
      await loadSounds();
    }

    if (audioContext.state === 'suspended') {
      await audioContext.resume();
    }

    const buffer = audioBufferMap.get(tone.id);
    if (!buffer) {
      console.warn(`Sound ${tone.name} not loaded yet`);
      return;
    }

    const source = audioContext.createBufferSource();
    source.buffer = buffer;

    const gainNode = audioContext.createGain();
    gainNode.gain.value = volume / 100;

    source.connect(gainNode);
    gainNode.connect(audioContext.destination);

    source.onended = () => {
      if (currentSource === source) {
        currentSource = null;
        currentGainNode = null; 
      }
    };

    currentSource = source;
    currentGainNode = gainNode; 
    source.start();
  }

  async function handleFileSelect() {
    try {
      const result = await Dialogs.OpenFile({
        Title: "Select a sound file",
        Filters: [
          {
            DisplayName: "Audio Files",
            Pattern: "*.mp3;*.wav;*.ogg"
          }
        ],
        AllowsMultipleSelection: false
      });

      if (result) {
        const path = Array.isArray(result) ? result[0] : result;
        const base64Data = await FileReading.ReadFile(path);
        const binaryString = window.atob(base64Data);
        const bytes = new Uint8Array(binaryString.length);
        for (let i = 0; i < binaryString.length; i++) {
          bytes[i] = binaryString.charCodeAt(i);
        }

        const buffer = await audioContext.decodeAudioData(bytes.buffer);
        const newTone = {
          id: 'custom-' + Date.now(),
          name: path.split(/[\\/]/).pop() || 'Custom',
          url: path
        };

        audioBufferMap.set(newTone.id, buffer);
        customTone = newTone;
        selectedTone = newTone; 
      }
    } catch (err) {
      console.error("Selection error:", err);
    }
  }

  function removeCustomTone() {
    if (customTone) {
      if (currentSource && selectedTone === customTone) {
        try { currentSource.stop(); } catch (e) {}
        currentSource = null;
        currentGainNode = null;
      }
      audioBufferMap.delete(customTone.id);
      customTone = null;
      selectedTone = soundOptions[0];
    }
  }

  $effect(() => {
    if (currentGainNode) {
      currentGainNode.gain.value = volume / 100;
    }
  });

  $effect(() => {
    if (!isRunning && !isPaused) {
      leftSeconds = minutes * 60 + seconds;
    }
  });

  let formattedTime = $derived.by(() => {
    const mins = Math.floor(leftSeconds / 60);
    const secs = leftSeconds % 60;
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  });
</script>

<main>
  {#if isDisplay}
    <!-- DISPLAY MONITOR: crna pozadina -->
    <div class="flex h-screen w-full items-center justify-center bg-black overflow-hidden">
      <div class="text-center w-full px-4">
        {#if showMessage}
          <h1 class="text-8xl md:text-8xl font-bold tracking-tight animate-fade-in">
            {displayMessage}
          </h1>
        {:else}
          <h1 class="text-[35vw] font-medium leading-none tabular-nums tracking-tighter text-white">
            {displayTime}
          </h1>
        {/if}
      </div>
    </div>
  {:else}
    <div class="h-screen overflow-hidden flex flex-col px-5 py-4">

      <div class="grid grid-cols-2 gap-4 h-full items-start">

        <!-- LIJEVO: Trajanje → Poruka → Preview → Gumbi -->
        <div class="flex flex-col gap-4">

          <section class="bg-[#1f2229] px-5 py-4 rounded-2xl border border-gray-800 flex-shrink-0">
            <h2 class="text-xs font-bold text-blue-400 uppercase tracking-widest mb-3">Trajanje</h2>
            <div class="flex gap-4">
              <div class="flex-1">
                <p class="text-gray-400 text-[10px] uppercase mb-1 block">Min</p>
                <input type="number" bind:value={minutes} class="w-full bg-transparent text-center border-b rounded-xl border-gray-600 focus:border-blue-500 outline-none text-2xl py-1" />
              </div>
              <div class="flex-1">
                <p class="text-gray-400 text-[10px] uppercase mb-1 block">Sec</p>
                <input type="number" bind:value={seconds} class="w-full bg-transparent text-center border-b rounded-xl border-gray-600 focus:border-blue-500 outline-none text-2xl py-1" />
              </div>
            </div>
          </section>

          <section class="bg-[#1f2229] px-5 py-4 rounded-2xl border border-gray-800 flex-shrink-0">
            <h2 class="text-xs font-bold text-blue-400 uppercase tracking-widest mb-3">Poruka</h2>
            <input type="text" bind:value={message} class="w-full bg-[#2a2e37] rounded-xl px-3 py-2 border-b border-gray-500 focus:border-blue-500 outline-none text-sm" />
          </section>

          <!-- Preview odmah ispod Poruke -->
          <div class="bg-black rounded-2xl border border-gray-700 px-5 py-4 flex items-center justify-between flex-shrink-0">
            <div>
              <p class="text-[10px] font-bold text-blue-400 uppercase tracking-widest mb-1">Preview — Display Monitor</p>
              <span class="text-6xl font-medium tabular-nums tracking-tighter text-white leading-none">
                {formattedTime}
              </span>
            </div>
            <div class="flex items-center gap-2">
              {#if isRunning}
                <span class="w-2.5 h-2.5 rounded-full bg-green-400 animate-pulse"></span>
                <span class="text-xs text-green-400 uppercase tracking-wider">Live</span>
              {:else if isPaused}
                <span class="w-2.5 h-2.5 rounded-full bg-yellow-400"></span>
                <span class="text-xs text-yellow-400 uppercase tracking-wider">Pauza</span>
              {:else}
                <span class="w-2.5 h-2.5 rounded-full bg-gray-600"></span>
                <span class="text-xs text-gray-500 uppercase tracking-wider">Mirovanje</span>
              {/if}
            </div>
          </div>

          <!-- Gumbi ispod previewa -->
          <div class="flex justify-center items-center gap-3 flex-shrink-0">
            {#if !isRunning && !isPaused}
              <button
                onclick={handleStart}
                class="bg-blue-600 hover:bg-blue-700 border border-transparent hover:border-blue-500 text-white font-bold py-3.5 px-16 text-lg rounded-full transition-all active:scale-95 shadow-xl"
              >
                Start
              </button>
            {:else if isPaused}
              <button
                onclick={handleStart}
                class="bg-blue-600 hover:bg-blue-700 border border-transparent hover:border-blue-500 text-white font-bold py-3.5 px-16 rounded-full transition-all active:scale-95 shadow-xl flex items-center justify-center"
              >
                <img src="/play-svgrepo-com.svg" class="w-6 h-6" alt="Play">
              </button>
              <button
                onclick={handleCancel}
                class="bg-[#2a2e37] hover:bg-gray-700 text-gray-400 font-bold py-3.5 px-16 rounded-full transition-all border border-gray-600 flex items-center justify-center"
              >
                <img src="/arrow-rotate-left-svgrepo-com.svg" alt="Reset" class="w-6 h-6">
              </button>
            {:else}
              <button
                onclick={handlePause}
                class="bg-blue-600 hover:bg-blue-700 border border-transparent hover:border-blue-500 text-white font-bold py-3.5 px-16 rounded-full transition-all active:scale-95 shadow-xl flex items-center justify-center"
              >
                <img src="/pause-svgrepo-com.svg" alt="Pause" class="w-6 h-6">
              </button>
              <button
                onclick={handleCancel}
                class="bg-[#2a2e37] hover:bg-gray-700 text-gray-400 font-bold py-3.5 px-16 rounded-full transition-all border border-gray-600 flex items-center justify-center"
              >
                <img src="/arrow-rotate-left-svgrepo-com.svg" alt="Reset" class="w-6 h-6">
              </button>
            {/if}
          </div>

        </div>

        <!-- DESNO: Tonovi + Volume -->
        <section class="bg-[#1f2229] px-5 py-4 rounded-2xl border border-gray-800 flex flex-col">
          <h2 class="text-xs font-bold text-blue-400 uppercase tracking-widest mb-3">Izbor tona</h2>

          <div class="flex flex-col gap-1.5 flex-1">
            <div class="grid grid-cols-1 gap-1.5">
              {#each soundOptions as tone}
                <button
                  onclick={() => { playSound(tone); selectedTone = tone; }}
                  class="px-4 py-2 text-sm rounded-xl text-left transition-all flex items-center justify-between group
                    {selectedTone?.id === tone.id
                      ? 'bg-blue-600 text-white shadow-lg shadow-blue-900/40'
                      : 'bg-[#2a2e37] text-gray-400 hover:bg-gray-700 border border-transparent hover:border-gray-600'}"
                >
                  <span class={selectedTone?.id === tone.id ? 'font-bold' : 'font-medium'}>{tone.name}</span>
                  {#if selectedTone?.id === tone.id}
                    <img src="/volume-low-svgrepo-com.svg" class="w-4 h-4" alt="Volume-Icon" />
                  {/if}
                </button>
              {/each}
            </div>

            {#if customTone}
              <div class="flex items-center gap-1.5">
                <button
                  onclick={() => { if (customTone) { playSound(customTone); selectedTone = customTone; } }}
                  class="flex-1 px-4 py-2 text-sm rounded-xl text-left transition-all flex items-center justify-between gap-2
                    {selectedTone?.id === customTone?.id
                      ? 'bg-blue-600 text-white shadow-lg shadow-blue-900/40'
                      : 'bg-[#2a2e37] text-gray-400 hover:bg-gray-700 border border-transparent hover:border-gray-600'}"
                >
                  <span class="truncate max-w-[90px] {selectedTone?.id === customTone?.id ? 'font-bold' : 'font-medium'}">{customTone.name}</span>
                  {#if selectedTone?.id === customTone?.id}
                    <img src="/volume-low-svgrepo-com.svg" class="w-4 h-4 flex-shrink-0" alt="Volume-Icon" />
                  {/if}
                </button>
                <button
                  onclick={removeCustomTone}
                  class="p-2 bg-[#2a2e37] hover:bg-red-500/20 rounded-xl border border-transparent hover:border-gray-600 flex-shrink-0"
                >
                  <img src="/x-1-svgrepo-com.svg" class="w-4 h-4" alt="Remove">
                </button>
              </div>
            {:else}
              <button
                onclick={handleFileSelect}
                class="px-4 py-2 text-sm rounded-xl transition-all flex items-center justify-center bg-[#2a2e37] text-gray-400 hover:bg-gray-700 border border-transparent hover:border-gray-600 w-full gap-2"
              >
                <span class="font-medium">Izaberi ton</span>
                <img src="/files-svgrepo-com.svg" class="w-4 h-4" alt="File">
              </button>
            {/if}
          </div>

          <!-- Volume pri dnu -->
          <div class="mt-3 flex-shrink-0">
            <div class="flex justify-between items-center mb-1.5">
              <span class="text-xs text-gray-400 uppercase">Master Volume</span>
              <span class="text-xs text-gray-400">{volume}%</span>
            </div>
            <input
              type="range"
              bind:value={volume}
              min="0"
              max="100"
              class="w-full accent-blue-600 h-2 bg-gray-700 rounded-lg cursor-pointer"
            />
          </div>
        </section>

      </div>
    </div>
  {/if}
</main>

<style>
  :global(body) {
    margin: 0;
    background-color: #1b2636;
    color: white;
    overflow: hidden;
  }

  /* Sakriva strelice za Chrome, Safari, Edge */
  input::-webkit-outer-spin-button,
  input::-webkit-inner-spin-button {
    -webkit-appearance: none;
    margin: 0;
  }

  /* Sakriva strelice za Firefox */
  input[type='number'] {
    -moz-appearance: textfield;
    appearance: textfield;
  }
</style>
