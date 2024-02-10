<script lang="ts">
    import Snippet from "./Snippet.svelte";

    async function load() {
        return await (await fetch(`http://localhost:3000/api/lint-highlights?lintId=${lintId}`)).json()
    }

    async function moderate(path: string, startLine: number, endLine: number, status: string) {
        await fetch(`http://localhost:3000/api/lint-highlights/moderate?lintId=${lintId}&path=${path}&startLine=${startLine}&endLine=${endLine}&status=${status}`)
    }

    export let lintId: string
</script>

<h2 style="text-align: left">highlights</h2>
{#await load()}
    <p>loading</p>
{:then items}
    {#each items as item}
        <div>
            <div>
                <div>
                    <a href={`${item.repo.gitUrl}/blob/${item.repo.gitBranch}/${item.path}#L${item.startLine}-L${item.endLine}`} target="_blank">
                        {item.path}#L{item.startLine}-L{item.endLine}
                    </a>
                    {#if item.status === "pending"}
                        <button class="accept" on:click={() => moderate(item.path, item.startLine, item.endLine, "accepted")}>approve bug</button>
                        <button class="reject" on:click={() => moderate(item.path, item.startLine, item.endLine, "rejected")}>reject bug</button>
                    {/if}
                    {#if item.status === "accepted"}
                        <div class="accept">accepted</div>
                    {/if}
                    {#if item.status === "rejected"}
                        <div class="reject">rejected</div>
                    {/if}
                </div>
                <div class="explanation">
                    {item.explanation}
                </div>
            </div>
            <Snippet snippet={item.snippet}/>
        </div>
    {/each}
{:catch error}
    <p>{error}</p>
{/await}

<style>
    div.accept {
        font-weight: bold;
        display: inline-block;
        padding: 0.6em 1.2em;
    }

    div.reject {
        font-weight: bold;
        display: inline-block;
        padding: 0.6em 1.2em;
    }

    .accept {
        color: green;
    }

    .reject {
        color: orangered;
    }

    .explanation {
        font-style: italic;
    }
</style>