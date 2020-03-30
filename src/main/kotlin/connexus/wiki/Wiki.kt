package connexus.wiki

import mu.KotlinLogging
import org.commonmark.Extension
import org.commonmark.ext.gfm.strikethrough.StrikethroughExtension
import org.commonmark.ext.gfm.tables.TablesExtension
import org.commonmark.node.Node
import org.commonmark.parser.Parser
import org.commonmark.renderer.html.HtmlRenderer
import java.io.File
import java.lang.Exception
import java.util.*
import kotlin.IllegalStateException

private val logger = KotlinLogging.logger {}

data class Wiki(val rootFolder: File, val homeTopic: String) {

    private val extensions: List<Extension> = Arrays.asList(
            TablesExtension.create(),
            StrikethroughExtension.create())

    private val parser: Parser = Parser
            .builder()
            .extensions(extensions)
            .build()

    private val renderer = HtmlRenderer
            .builder()
            .extensions(extensions)
            .build()

    fun renderHtml(contents: String): String {
        val document: Node = parser.parse(contents)
        return renderer.render(document)
    }

    fun renderToFile(contents: String, fileName: String) {
        val opHtml = renderHtml(contents)
        File(fileName).writeText(PRE + opHtml + POST)
    }

    private fun pagePath(title: String): File {
        val parts = title.split("/")
        val nParts = parts.count()
        if (nParts > 1) {
            // if this is a directory rather than a file name,
            // ensure that the directory exists
            val dirPath = File(rootFolder.path
                    + "/"
                    + parts
                    .take(nParts - 1)
                    .joinToString("/"))

            val dirExists = try {
                if (dirPath.exists()) {
                    true
                } else {
                    dirPath.mkdir()
                }
            } catch (e: Exception) {
                throw IllegalStateException(
                        "error creating directory:${dirPath} ${e.message}")
            }

            if (!dirExists) {
                logger.error { "failed to create directory: $dirPath" }
                throw IllegalStateException(
                        "failed to create directory:${dirPath}")
            }
        }
        return rootFolder.resolve("$title.md")
    }

    private fun loadMarkdown(title: String): Page {
        val mdFile = pagePath(title)
        val content = try {
            mdFile.readText()
        } catch (e: Exception) {
            throw IllegalStateException(
                    "error reading ${mdFile.path}: ${e.message}")
        }
        return Page(title, content)
    }

    fun savePage(p: Page) {
        val mdFile = pagePath(p.title) ?: return
        logger.info { "writing to ${mdFile.path}" }
        try {
            mdFile.writeText(p.body)
        } catch (e: Exception) {
            throw IllegalStateException(
                    "error writing to page: ${p.title} - ${e.message}")
        }
    }

    fun loadPage(path: String): String {
        return try {
            val page = loadMarkdown(path)
            renderHtml(page.body)
        } catch (e: Exception) {
            "<p>Failed to load page: $path - ${e.message}</p>"
        }
    }
}

data class Page(val title: String, val body: String)

private const val PRE = """
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
<title></title>
<meta charset="utf-8" />
<link rel="stylesheet" type="text/css" href="style.css" />
</head>
<body>
"""

private const val POST = """
</body>
</html>
"""
