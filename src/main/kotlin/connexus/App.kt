/*
 * This Kotlin source file was generated by the Gradle 'init' task.
 */
package connexus

import connexus.wiki.Wiki
import io.ktor.application.call
import io.ktor.html.respondHtml
import io.ktor.http.ContentType
import io.ktor.response.respondRedirect
import io.ktor.response.respondText
import io.ktor.routing.get
import io.ktor.routing.routing
import io.ktor.server.engine.embeddedServer
import io.ktor.server.netty.Netty
import kotlinx.html.*
import org.commonmark.Extension
import org.commonmark.ext.gfm.strikethrough.StrikethroughExtension
import org.commonmark.ext.gfm.tables.TablesExtension
import org.commonmark.node.Node
import org.commonmark.parser.Parser
import org.commonmark.renderer.html.HtmlRenderer
import java.io.File
import java.util.*


fun main(args: Array<String>) {

    if (args.count() != 2) {
        println("please provide root folder and home page title")
        return
    }

    val rootFolder = File(args[0].replaceFirst("^~".toRegex(),
            System.getProperty("user.home")))

    val wiki = Wiki(rootFolder = rootFolder, homeTopic = args[1])

    val server = embeddedServer(Netty, port = 8080) {
        routing {
            get("/") {
                call.respondRedirect("/view/${wiki.homeTopic}")
            }
            get("/view/{path...}") {
                val path = call.parameters["path"] ?: return@get
                call.respondHtml {
                    head {
                        title {
                            +path
                        }
                        meta(charset = "utf-8")
                        link(rel = "stylesheet",
                                type = "text/css",
                                href = "style.css")
                    }
                    body {
                        val content = wiki.loadPage(path)
                        unsafe {
                            +content
                        }
                    }
                }
            }
        }
    }
    server.start(wait = true)
}

fun main() {
    val home = System.getProperty("user.home")
    val wikiDir = File("$home/tmp/testwiki")
    val input = wikiDir.resolve("full-feature-testing.md")
    val contents = input.readText()

    val wiki = Wiki(wikiDir, "test1.md")
    wiki.renderToFile(contents, "test.html")

}
