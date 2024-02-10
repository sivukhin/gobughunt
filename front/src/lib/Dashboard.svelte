<script>
    import Repos from "./Repos.svelte";
    import Linters from "./Linters.svelte";
</script>

<div class="overview">
    <div>
        <Linters/>
    </div>
    <div>
        <Repos/>
    </div>
</div>

<style>
    .overview {
        display: flex;
        flex-direction: row;
        justify-content: space-between;
    }
</style>