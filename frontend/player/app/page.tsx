export default function Home() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-24">
      <div className="z-10 max-w-5xl w-full items-center justify-between font-mono text-sm lg:flex">
        <p className="fixed left-0 top-0 flex w-full justify-center border-b border-gray-300 bg-gradient-to-b from-zinc-200 pb-6 pt-8 backdrop-blur-2xl dark:border-neutral-800 dark:bg-zinc-800/30 dark:from-inherit lg:static lg:w-auto lg:rounded-xl lg:border lg:bg-gray-200 lg:p-4 lg:dark:bg-zinc-800/30">
          ProveMySelf Player
        </p>
      </div>

      <div className="relative flex place-items-center">
        <h1 className="text-6xl font-bold text-center">
          Quiz Player
        </h1>
      </div>

      <div className="mt-16 text-center">
        <p className="text-xl text-gray-600 dark:text-gray-400">
          Interactive quiz experiences powered by Adaptive Cards
        </p>
        <p className="mt-4 text-sm text-gray-500 dark:text-gray-500">
          Load a quiz bundle to start playing
        </p>
      </div>
    </main>
  );
}