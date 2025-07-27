export class BaseRequest {
    constructor(name, data) {
        this.name = name
        this.data = data
    }

    static fromJson(json) {
        return JSON.parse(json)
    }

    toJson() {
        return JSON.stringify(this)
    }
}

export class BaseResponse {
    constructor(code, message, data) {
        this.code = code
        this.message = message
        this.data = data
    }

    static fromJson(json) {
        return JSON.parse(json)
    }

    toJson() {
        return JSON.stringify(this)
    }
}