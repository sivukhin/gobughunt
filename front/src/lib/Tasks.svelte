<script lang="ts">
    import {navigate} from "svelte-routing";

    async function load(skip: number, take: number) {
        return await (await fetch(`http://localhost:3000/api/lint-tasks?skip=${skip}&take=${take}`)).json()
    }
</script>

<h2 style="text-align: left">lint tasks</h2>
{#await load(0, 10)}
    <p>loading</p>
{:then items}
    <table>
        <tr>
            <th style="text-align: left">linter</th>
            <th style="text-align: left">repo</th>
            <th style="text-align: left">status</th>
            <th style="text-align: left">duration</th>
        </tr>
        {#each items as item}
            <tr class="link" on:click={() => navigate(`/highlights/${item.id}`)}>
                <td style="text-align: left">{item.linter.id}</td>
                <td style="text-align: left">{item.repo.id}</td>
                <td style="text-align: left" class={`${item.status} highlight`}>
                    {item.status}
                    {#if item.statusComment !== ""}
                        (<span class="comment" title={item.statusComment}>{item.statusComment}</span>)
                    {/if}
                </td>
                <td style="text-align: left">{`${item.lintDurationSec.toFixed(2)}`} sec.</td>
            </tr>
        {/each}
    </table>
{:catch error}
    <p>{error}</p>
{/await}


<style>
    .succeed.highlight {
        background-color: lightgreen;
    }

    .failed.highlight {
        background-color: lightcoral;
    }

    .link:hover {
        cursor: pointer;
        background-color: lightblue;
    }

    .comment {
        vertical-align: bottom;
        display: inline-block;
        max-width: 100px;
        overflow: hidden;
        white-space: nowrap;
        text-overflow: ellipsis;
    }

    table {
        border-spacing: 0;
        border-collapse: collapse;
    }

    th {
        border-bottom: 2px solid dimgray;
    }

    td, th {
        padding: 4px;
        border-left: 1px solid dimgray;
        border-right: 1px solid dimgray;
    }
</style>