(function (root, factory) {
    if (typeof define === 'function' && define.amd) {
        define([], factory);
    } else if (typeof module === 'object' && module.exports) {
        module.exports = factory();
    } else {
        root.XGDNPay = factory();
    }
}(typeof self !== 'undefined' ? self : this, function () {
    'use strict';

    var DEFAULT_BASE_URL = 'https://pay.xgdn.net';
    var SSE_PATH = '/api/v1/sse/subscribe/';
    var PAY_INFO_PATH = '/api/v1/pay/';
    var JSAPI_PAY_PATH = '/jsapi';
    var DEFAULT_TIMEOUT = 300000;
    var RECONNECT_BASE_DELAY = 1000;
    var RECONNECT_MAX_DELAY = 30000;
    var RECONNECT_MULTIPLIER = 2;

    function XGDNPay(options) {
        if (!(this instanceof XGDNPay)) {
            return new XGDNPay(options);
        }

        options = options || {};

        this.baseURL = options.baseURL || DEFAULT_BASE_URL;
        this.appID = options.appID || '';
    }

    XGDNPay.prototype.getPayInfo = function (orderNo) {
        if (!orderNo || typeof orderNo !== 'string') {
            return Promise.reject(new Error('orderNo 是必填参数'));
        }

        var url = this.baseURL + PAY_INFO_PATH + encodeURIComponent(orderNo);

        return fetch(url, {
            method: 'GET',
            headers: { 'Accept': 'application/json' }
        }).then(function (response) {
            if (!response.ok) {
                throw new Error('HTTP ' + response.status + ': 获取支付信息失败');
            }
            return response.json();
        }).then(function (result) {
            if (result.code !== 0) {
                throw new Error(result.message || '获取支付信息失败');
            }
            return result.data;
        });
    };

    XGDNPay.prototype.getJSAPIPayInfo = function (orderNo, openID) {
        if (!orderNo || typeof orderNo !== 'string') {
            return Promise.reject(new Error('orderNo 是必填参数'));
        }

        var url = this.baseURL + PAY_INFO_PATH + encodeURIComponent(orderNo) + JSAPI_PAY_PATH;

        if (openID) {
            url += '?openid=' + encodeURIComponent(openID);
        }

        return fetch(url, {
            method: 'GET',
            headers: { 'Accept': 'application/json' }
        }).then(function (response) {
            if (!response.ok) {
                throw new Error('HTTP ' + response.status + ': 获取JSAPI支付信息失败');
            }
            return response.json();
        }).then(function (result) {
            if (result.code !== 0) {
                throw new Error(result.message || '获取JSAPI支付信息失败');
            }
            return result.data;
        });
    };

    XGDNPay.prototype.showQRCode = function (container, codeURL, options) {
        if (!container) {
            console.error('[XGDNPay] container 不能为空');
            return null;
        }

        if (!codeURL) {
            console.error('[XGDNPay] codeURL 不能为空');
            return null;
        }

        options = options || {};
        var size = options.size || 200;
        var margin = options.margin || 2;
        var colorDark = options.colorDark || '#000000';
        var colorLight = options.colorLight || '#ffffff';

        var el = typeof container === 'string'
            ? document.querySelector(container)
            : container;

        if (!el) {
            console.error('[XGDNPay] 找不到容器元素:', container);
            return null;
        }

        el.innerHTML = '';

        var img = document.createElement('img');
        img.style.width = size + 'px';
        img.style.height = size + 'px';
        img.alt = '支付二维码';
        img.src = codeURL;

        if (options.onLoad && typeof options.onLoad === 'function') {
            img.onload = function () { options.onLoad(img); };
        }
        if (options.onError && typeof options.onError === 'function') {
            img.onerror = function () { options.onError(new Error('二维码加载失败')); };
        }

        el.appendChild(img);

        return img;
    };

    XGDNPay.prototype.createPayment = function (orderNo, options) {
        var self = this;
        options = options || {};

        var result = {
            orderNo: orderNo,
            _watcher: null,
            _qrImage: null,

            start: function () {
                return self.getPayInfo(orderNo).then(function (payInfo) {
                    result.payInfo = payInfo;

                    if (payInfo.need_auth) {
                        if (options.onNeedAuth) {
                            options.onNeedAuth(payInfo);
                        }
                        return result;
                    }

                    if (payInfo.pay_url || payInfo.code_url || payInfo.qrcode_url) {
                        if (options.container && (payInfo.code_url || payInfo.qrcode_url)) {
                            result._qrImage = self.showQRCode(
                                options.container,
                                payInfo.qrcode_url || payInfo.code_url,
                                {
                                    size: options.qrSize || 200,
                                    onLoad: options.onQRCodeLoad,
                                    onError: options.onQRCodeError
                                }
                            );
                        }

                        if (options.onReady) {
                            options.onReady(payInfo);
                        }

                        result._watcher = XGDNPay.watch(orderNo, {
                            baseURL: self.baseURL,
                            timeout: options.timeout || DEFAULT_TIMEOUT,
                            onConnected: options.onConnected,
                            onPaid: function (data) {
                                result._watcher = null;
                                if (options.onPaid) {
                                    options.onPaid(data);
                                }
                            },
                            onTimeout: function () {
                                result._watcher = null;
                                if (options.onTimeout) {
                                    options.onTimeout();
                                }
                            },
                            onError: function (err) {
                                if (options.onError) {
                                    options.onError(err);
                                }
                            }
                        });
                    }

                    return result;
                });
            },

            close: function () {
                if (result._watcher) {
                    result._watcher.close();
                    result._watcher = null;
                }
                if (options.container) {
                    var el = typeof options.container === 'string'
                        ? document.querySelector(options.container)
                        : options.container;
                    if (el) {
                        el.innerHTML = '';
                    }
                }
            }
        };

        return result;
    };

    XGDNPay.watch = function (orderNo, options) {
        if (!orderNo || typeof orderNo !== 'string') {
            throw new Error('orderNo 是必填参数');
        }

        var opts = mergeWatchOptions(options);
        return new PaymentWatcher(orderNo, opts);
    };

    function mergeWatchOptions(options) {
        if (!options || typeof options !== 'object') {
            options = {};
        }

        return {
            baseURL: options.baseURL || DEFAULT_BASE_URL,
            timeout: options.timeout || DEFAULT_TIMEOUT,
            onConnected: typeof options.onConnected === 'function' ? options.onConnected : noop,
            onPaid: typeof options.onPaid === 'function' ? options.onPaid : noop,
            onTimeout: typeof options.onTimeout === 'function' ? options.onTimeout : noop,
            onError: typeof options.onError === 'function' ? options.onError : noop,
        };
    }

    function PaymentWatcher(orderNo, options) {
        this._orderNo = orderNo;
        this._options = options;
        this._eventSource = null;
        this._timeoutTimer = null;
        this._reconnectTimer = null;
        this._reconnectDelay = RECONNECT_BASE_DELAY;
        this._closed = false;
        this._connected = false;

        this._startTimeout();
        this._connect();
    }

    PaymentWatcher.prototype.close = function () {
        this._closed = true;
        this._clearReconnect();
        this._disconnectSSE();
        this._clearTimeout();
    };

    PaymentWatcher.prototype._connect = function () {
        if (this._closed) return;

        var url = this._options.baseURL + SSE_PATH + this._orderNo;

        try {
            this._eventSource = new EventSource(url);
        } catch (err) {
            this._scheduleReconnect();
            return;
        }

        var self = this;

        this._eventSource.onopen = function () {
            self._connected = true;
            self._reconnectDelay = RECONNECT_BASE_DELAY;
            self._options.onConnected();
        };

        this._eventSource.addEventListener('connected', function () {
            self._connected = true;
        });

        this._eventSource.addEventListener('message', function (e) {
            self._handleMessage(e);
        });

        this._eventSource.onerror = function () {
            self._connected = false;
            self._disconnectSSE();
            self._scheduleReconnect();
        };
    };

    PaymentWatcher.prototype._handleMessage = function (e) {
        try {
            var data = JSON.parse(e.data);

            if (data.status === 'paid') {
                this.close();
                this._options.onPaid(data);
            }
        } catch (err) {
        }
    };

    PaymentWatcher.prototype._disconnectSSE = function () {
        if (this._eventSource) {
            this._eventSource.onopen = null;
            this._eventSource.onerror = null;
            this._eventSource.close();
            this._eventSource = null;
        }
        this._connected = false;
    };

    PaymentWatcher.prototype._scheduleReconnect = function () {
        if (this._closed) return;

        var self = this;
        var delay = this._reconnectDelay;

        this._reconnectDelay = Math.min(
            this._reconnectDelay * RECONNECT_MULTIPLIER,
            RECONNECT_MAX_DELAY
        );

        this._reconnectTimer = setTimeout(function () {
            self._reconnectTimer = null;
            self._connect();
        }, delay);
    };

    PaymentWatcher.prototype._clearReconnect = function () {
        if (this._reconnectTimer) {
            clearTimeout(this._reconnectTimer);
            this._reconnectTimer = null;
        }
    };

    PaymentWatcher.prototype._startTimeout = function () {
        var self = this;
        this._timeoutTimer = setTimeout(function () {
            self._timeoutTimer = null;
            self.close();
            self._options.onTimeout();
        }, this._options.timeout);
    };

    PaymentWatcher.prototype._clearTimeout = function () {
        if (this._timeoutTimer) {
            clearTimeout(this._timeoutTimer);
            this._timeoutTimer = null;
        }
    };

    function noop() {}

    XGDNPay.DEFAULT_BASE_URL = DEFAULT_BASE_URL;
    XGDNPay.DEFAULT_TIMEOUT = DEFAULT_TIMEOUT;
    XGDNPay.PayType = {
        NATIVE: 'native',
        JSAPI: 'jsapi'
    };
    XGDNPay.OrderStatus = {
        PENDING: 0,
        PAID: 1,
        CLOSED: 2,
        REFUNDED: 3
    };
    XGDNPay.RefundStatus = {
        PROCESSING: 0,
        SUCCESS: 1,
        CLOSED: 2,
        FAILED: 3,
        ABNORMAL: 4
    };

    return XGDNPay;
}));
