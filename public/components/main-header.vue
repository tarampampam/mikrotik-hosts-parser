<template>
    <div class="headline">
        <h1 class="logo text-white">
            <span class="logo"><i class="fas fa-bolt"></i></span>
            <span :title="version">MikroTik StopAD</span>
            <small>Сделаем рекламы чуть меньше</small>
        </h1>
    </div>
</template>

<script>
    /* global module */

    module.exports = {
        props: {
            version: {
                default: undefined,
                type: String
            },
        },
    };
</script>

<style scoped>
    @-webkit-keyframes heartbeat {
        0% {
            -webkit-transform: scale(1)
        }
        3% {
            -webkit-transform: scale(1.1)
        }
        6% {
            -webkit-transform: scale(1)
        }
        9% {
            -webkit-transform: scale(1.1)
        }
        12% {
            -webkit-transform: scale(1)
        }
        100% {
            -webkit-transform: scale(1)
        }
    }

    .headline {
        padding: 73px 0 110px;
        text-align: center;
    }

    .headline .logo {
        letter-spacing: -2px;
        margin: 10px 0;
        font-size: 4em;
    }

    .headline .logo .logo {
        overflow: hidden;
        font-size: 2em;
        display: block;
        color: #16a085;
        animation: heartbeat 4500ms infinite linear;
    }

    .headline .logo small {
        display: block;
        font-size: .35em;
        opacity: .6;
        font-weight: 200;
        letter-spacing: 0;
    }
</style>
