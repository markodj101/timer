<script lang="ts">
  import './app.css';
  import { Events } from "@wailsio/runtime";
  import { GreetService } from "../bindings/changeme";
  import {Button, Range, Label} from "flowbite-svelte";
 
  import { onMount } from 'svelte';
	let name: string = '';
	let result: string = 'Please enter your name below 👇';
	let time: string = 'Listening for Time event...';
	let path = window.location.pathname;

  const soundOptions = [
    {id: 'Ton 1', name: 'Ton 1' ,url:'/1.mp3'},
    {id: 'Ton 2', name: 'Ton 2',url:'/2.mp3'},
    {id: 'Ton 3', name: 'Ton 3',url:'/3.mp3'},
    {id: 'Ton 4', name: 'Ton 4',url:'/4.mp3'}, 
    {id: 'Ton 5', name: 'Ton 5',url:'/5.mp3'}   
  ];

  let hours = 0, minutes = 5, seconds = 0;
  let message = "Time is up!";
  let volume = 80;
  let selectedTone = soundOptions[0];
  let timeLeft = "00:05:00";
  let isRunning = false;
  let audioContext : AudioContext;
  let audioBufferMap = new Map<string, AudioBuffer>();


  function handleStart(){
    isRunning = true;
    // Ovdje bi išao kod za pokretanje tajmera
  }

  function handleCancel(){
    isRunning = false;
    // Ovdje bi išao kod za zaustavljanje tajmera
  }


  onMount(async ()=>{
    audioContext = new AudioContext();
    await loadSounds();
  })

  async function loadSounds(){
    for(let tone of soundOptions){
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

    source.start();
  }

  function onTimerComplete(){
    playSound(selectedTone);
  }
	
</script>

<main>
	{#if path === '/display'}
		<div class="fullscreen-timer"><h1>00:00</h1></div>
	{:else}
		<div class="controls min-h-screen flex flex-col justify-center p-6">
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
      on:click={() => { playSound(tone); selectedTone = tone; }}
      class="p-4 text-sm rounded-2xl text-left transition-all flex items-center justify-between group
      {selectedTone === tone 
        ? 'bg-blue-600 text-white shadow-lg shadow-blue-900/40' 
        : 'bg-[#2a2e37] text-gray-400 hover:bg-gray-700 border border-transparent hover:border-gray-600'}"
    >
      <span class={selectedTone === tone ? 'font-bold' : 'font-medium'}>{tone.name}</span>
      
      {#if selectedTone === tone}
        <img src="/volume-low-svgrepo-com.svg" class=" w-5 h-5" />
      {/if}
    </button>
  {/each}
</div>
<button class="p-4 text-sm rounded-2xl text-left transition-all flex items-center justify-between bg-[#2a2e37] text-gray-400 hover:bg-gray-700 group border border-transparent hover:border-gray-600">
    <span class="font-medium flex items-center justify-center gap-2">
    Izaberi ton 
    <img src="/files-svgrepo-com.svg" class="w-5 h-5" alt="File">
  </span>
  </button>
      </div>

      <div class="mt-6">
        <div class="flex justify-between items-center mb-2">
          <div class="flex items-center gap-2 text-[13px] text-gray-400 uppercase">
            Master Volume
          </div>
          
          <span class="text-[13px] text-gray-400 flex items-center gap-1.5">{volume}%
 
  
</span>
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
  
  {#if !isRunning}
    <button
      on:click={handleStart}
      class="bg-blue-600 hover:bg-blue-700 border border-transparent hover:border-blue-500 text-white font-bold py-4 px-20 text-xl rounded-full transition-all active:scale-95 shadow-xl"
    >
      Start
    </button>
  {:else}
    <button
      on:click={handleCancel}
      class="bg-[#2a2e37] hover:bg-gray-700 text-gray-400 font-bold py-4 px-20 text-xl rounded-full transition-all border border-gray-600"
    >
      <img src="/arrow-u-up-left-svgrepo-com.svg" alt="Reset" class=" w-6 h-6">
    </button>

    <button
      class="bg-blue-600 hover:bg-blue-700 border border-transparent hover:border-blue-500 text-white font-bold py-4 px-20 text-xl rounded-full transition-all active:scale-95 shadow-xl"
    >
      <img src="/pause-svgrepo-com.svg" alt="Pause" class=" w-6 h-6">
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

	.fullscreen-timer {
		display: flex;
		justify-content: center;
		align-items: center;
		height: 100vh;
		font-size: 10rem;
	}
  /* Sakriva strelice za Chrome, Safari, Edge (Wails koristi Edge) */
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
