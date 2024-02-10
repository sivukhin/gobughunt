<script lang="ts">
    async function load() {
        return await (await fetch("http://localhost:3000/api/linters")).json()
    }
</script>

<h2 style="text-align: left">linters</h2>
{#await load()}
    <p>loading</p>
{:then items}
    <table>
        <tr>
            <th style="text-align: left">total</th>
            <th style="text-align: right">total</th>
            <th style="text-align: right">pending</th>
            <th style="text-align: right">accepted</th>
            <th style="text-align: right">rejected</th>
        </tr>
        {#each items as item}
            <tr>
                <td style="text-align: left"><a href="{item.gitUrl}" target="_blank">{item.id}</a></td>
                <td style="text-align: right">{item.totalHighlightDedup}</td>
                <td style="text-align: right">{item.pendingHighlightDedup}</td>
                <td style="text-align: right">{item.acceptedHighlightDedup}</td>
                <td style="text-align: right">{item.rejectedHighlightDedup}</td>
            </tr>
        {/each}
    </table>
{:catch error}
    <p>{error}</p>
{/await}

<style>
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