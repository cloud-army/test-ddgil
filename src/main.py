from flask import Flask, jsonify

app = Flask(__name__)


@app.get("/health")
def health():
    return jsonify({"status": "ok", "app": "test-ddgil"})


@app.get("/")
def root():
    return jsonify({"app": "test-ddgil", "version": "1.0.0"})


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8080)
