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

  onMount(async () => {
    audioContext = new AudioContext();
    await loadSounds();

    const blockRightClick = (e: MouseEvent) => e.preventDefault();
    window.addEventListener('contextmenu', blockRightClick);

     if (isDisplay) {
      Events.On("timer-update", (ev) => {
        showMessage = false;
        displayTime = ev.data;
      });

      Events.On("timer-complete", (ev) => {
        showMessage = true;
        displayMessage = ev.data;
      });
    }

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
    try {
      currentSource.stop();
    } catch (e) {}
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
        try {
          currentSource.stop();
        } catch (e) {}
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
    <div class="flex h-screen w-full items-center justify-center bg-[#1b2636] overflow-hidden">
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
    <div class="controls min-h-screen flex flex-col items-center p-6 py-12 ">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-2xl mx-auto">

       
        <div class="space-y-6">
          <section class="bg-[#1f2229] p-6 rounded-3xl border border-gray-800">
            <h2 class="text-xs font-bold text-blue-400 uppercase tracking-widest mb-4">Trajanje</h2>
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

          <section class="bg-[#1f2229] p-6 rounded-3xl border border-gray-800">
            <h2 class="text-xs font-bold text-blue-400 uppercase tracking-widest mb-4">Poruka</h2>
            <input type="text" bind:value={message} class="w-full bg-[#2a2e37] rounded-xl p-3 border-b border-gray-500 focus:border-blue-500 outline-none" />
          </section>
        </div>

      
        <section class="bg-[#1f2229] p-6 rounded-3xl border border-gray-800 flex flex-col">
          <h2 class="text-xs font-bold text-blue-400 uppercase tracking-widest mb-4">Izbor tona</h2>

          <div class="space-y-2 flex-1">
           
            <div class="grid grid-cols-1 gap-2">
              {#each soundOptions as tone}
                <button
                  onclick={() => { playSound(tone); selectedTone = tone; }}
                  class="p-4 text-sm rounded-2xl text-left transition-all flex items-center justify-between group
  {selectedTone?.id === tone.id
    ? 'bg-blue-600 text-white shadow-lg shadow-blue-900/40'
    : 'bg-[#2a2e37] text-gray-400 hover:bg-gray-700 border border-transparent hover:border-gray-600'}"
                >
                  <span class={selectedTone?.id === tone.id ? 'font-bold' : 'font-medium'}>{tone.name}</span>

                  {#if selectedTone?.id === tone.id}
                    <img src="/volume-low-svgrepo-com.svg" class="w-5 h-5" alt="Volume-Icon" />
                  {/if}
                </button>
              {/each}
            </div>

           
            {#if customTone}
              <div class="flex items-center gap-2">
                <button
                  onclick={() => { if (customTone) { playSound(customTone); selectedTone = customTone; } }}
                  class="p-4 text-sm rounded-2xl text-left transition-all flex items-center justify-between group gap-2
  {selectedTone?.id === customTone?.id
    ? 'bg-blue-600 text-white shadow-lg shadow-blue-900/40'
    : 'bg-[#2a2e37] text-gray-400 hover:bg-gray-700 border border-transparent hover:border-gray-600'}"
                >
                  <span class ={selectedTone?.id === customTone?.id ? 'font-bold' : 'font-medium'}>{customTone.name}</span>
                  {#if selectedTone?.id === customTone?.id}
                    <img src="/volume-low-svgrepo-com.svg" class="w-5 h-5" alt="Volume-Icon" />
                  {/if}
                </button>
                <button
                  onclick={removeCustomTone}
                  class="p-3 bg-[#2a2e37] hover:bg-red-500/20 rounded-2xl border border-transparent hover:border-gray-600"
                  
                >
                  <img src="/x-1-svgrepo-com.svg" class="w-4 h-4" alt="Remove">
        
                </button>
              </div>
            {:else}
              <button
                onclick={handleFileSelect}
                class="p-4 text-sm rounded-2xl text-left transition-all flex items-center justify-between bg-[#2a2e37] text-gray-400 hover:bg-gray-700 group border border-transparent hover:border-gray-600 w-full"
              >
                <span class="font-medium flex items-center justify-center gap-2 w-full">
                  Izaberi ton
                  <img src="/files-svgrepo-com.svg" class="w-5 h-5" alt="File">
                </span>
              </button>
            {/if}
          </div>

          <div class="mt-6">
            <div class="flex justify-between items-center mb-2">
              <div class="flex items-center gap-2 text-[13px] text-gray-400 uppercase">
                Master Volume
              </div>
              <span class="text-[13px] text-gray-400 flex items-center gap-1.5">{volume}%</span>
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

      
      <div class="mt-12 flex justify-center items-center w-full gap-4">
  {#if !isRunning && !isPaused}
    
    <button
      onclick={handleStart}
      class="bg-blue-600 hover:bg-blue-700 border border-transparent hover:border-blue-500 text-white font-bold py-4 px-20 text-xl rounded-full transition-all active:scale-95 shadow-xl"
    >
      Start
    </button>
  {:else if isPaused}
  
    <button
      onclick={handleStart}
      class="bg-blue-600 hover:bg-blue-700 border border-transparent hover:border-blue-500 text-white font-bold py-4 px-20 text-xl rounded-full transition-all active:scale-95 shadow-xl"
    >
      <img src="/play-svgrepo-com.svg" class="w-6 h-6" alt="Play">
    </button>
    <button
      onclick={handleCancel}
      class="bg-[#2a2e37] hover:bg-gray-700 text-gray-400 font-bold py-4 px-20 text-xl rounded-full transition-all border border-gray-600"
    >
      <img src="/arrow-rotate-left-svgrepo-com.svg" alt="Reset" class="w-6 h-6">
    </button>
  {:else}
    
    <button
      onclick={handlePause}
      class="bg-blue-600 hover:bg-blue-700 border border-transparent hover:border-blue-500 text-white font-bold py-4 px-20 text-xl rounded-full transition-all active:scale-95 shadow-xl"
    >
      <img src="/pause-svgrepo-com.svg" alt="Pause" class="w-6 h-6">
    </button>
    <button
      onclick={handleCancel}
      class="bg-[#2a2e37] hover:bg-gray-700 text-gray-400 font-bold py-4 px-20 text-xl rounded-full transition-all border border-gray-600"
    >
      <img src="/arrow-rotate-left-svgrepo-com.svg" alt="Reset" class="w-6 h-6">
    </button>
  {/if}
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